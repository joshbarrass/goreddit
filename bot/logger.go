package bot

import (
	"errors"
	"fmt"

	reddit "github.com/joshbarrass/goreddit/API"
	"github.com/sirupsen/logrus"
)

const logMessageMaxSubjectLength = 40
const defaultLogMessageTimeFormat = "Mon 2 Jan 2006, 15:04:05"

var defaultHookErrorLevels = []logrus.Level{logrus.PanicLevel, logrus.FatalLevel, logrus.ErrorLevel, logrus.WarnLevel}

// SendErrors sets up a hook to send errors via reddit to a username
func SendErrors(api *reddit.RedditAPI, username string, botName string) (*RedditErrorHook, error) {
	if username == "" {
		return nil, errors.New("no username to send errors to")
	}
	hook := RedditErrorHook{
		Reddit:     api,
		Username:   username,
		BotName:    botName,
		TimeFormat: defaultLogMessageTimeFormat,
		levels:     []logrus.Level{},
	}
	copy(defaultHookErrorLevels, hook.levels)
	logrus.AddHook(&hook)
	return &hook, nil
}

// RedditErrorHook is a hook for sending errors via reddit
type RedditErrorHook struct {
	Reddit     *reddit.RedditAPI
	Username   string
	BotName    string
	TimeFormat string
	levels     []logrus.Level
}

// Levels defines the levels that this hook will respond to
func (hook *RedditErrorHook) Levels() []logrus.Level {
	return hook.levels
}

// SetLevels allows for the firing error levels to be customised
func (hook *RedditErrorHook) SetLevels(levels []logrus.Level) {
	hook.levels = levels
}

// Fire is the function responsible for messaging the user
func (hook *RedditErrorHook) Fire(entry *logrus.Entry) error {
	// constuct message
	var level string
	switch entry.Level {
	case logrus.PanicLevel:
		level = "PANIC"
	case logrus.FatalLevel:
		level = "FATAL"
	case logrus.ErrorLevel:
		level = "ERROR"
	case logrus.WarnLevel:
		level = "WARN"
	}

	subject := fmt.Sprintf("%s in %s: %s", level, hook.BotName, entry.Message)

	// ensure subject is not longer than predefined maximum
	if len(subject) > logMessageMaxSubjectLength {
		subject = subject[:logMessageMaxSubjectLength] + "..."
	}

	message := fmt.Sprintf(
		`Time: %s
Message: %s
Data: %+v`, entry.Time.Format(hook.TimeFormat), entry.Message, entry.Data)

	// add caller data if available
	if entry.Caller != nil {
		message += fmt.Sprintf("\nCalling Function: '%s', '%s' line %d", entry.Caller.Function, entry.Caller.File, entry.Caller.Line)
	} else {
		message += "\nNo information available about the calling function."
	}

	// validate username has /u/ or /r/ at front, add /u/ if
	// necessary
	if hook.Username[:3] != "/u/" && hook.Username[:3] != "/r/" {
		hook.Username = fmt.Sprintf("/u/%s", hook.Username)
	}

	// send reddit message
	err := hook.Reddit.ComposeMessage(hook.Username, subject, message)
	if err != nil {
		return err
	}

	return nil
}
