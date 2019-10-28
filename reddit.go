package reddit

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// RedditAPI interfaces all the necessary functions to interact with
// Reddit via the official API
type RedditAPI struct {
	ClientID     string
	clientSecret string
	UserAgent    string
	Account      *RedditAccount
	Client       http.Client
	DebugMode    bool
}

// NewRedditAPI creates a new API with a given ClientID and
// ClientSecret and with an unauthenticated account
func NewRedditAPI(clientID, clientSecret, userAgent, username string, debugMode bool) *RedditAPI {
	reddit := RedditAPI{
		ClientID:     clientID,
		clientSecret: clientSecret,
		UserAgent:    userAgent,
		DebugMode:    debugMode,
	}
	account := RedditAccount{
		API:      &reddit,
		Username: username,
	}
	reddit.Account = &account

	return &reddit
}

func (api *RedditAPI) NewRequest(method string, u *url.URL, body io.Reader) (*http.Request, error) {
	// create new request
	url := u.String()
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	// set auth
	switch u.Host {
	case oauthHost:
		// if using OAUTH, check token is valid and set bearer
		// auth header
		if api.Account.Token.Token == "" {
			return nil, errors.New("no valid token")
		}
		if time.Now().After(api.Account.Token.Expiry) {
			return nil, errors.New("token expired")
		}
		req.Header.Set("Authorization", fmt.Sprintf("bearer %s", api.Account.Token.Token))
	case redditHost:
		// if using reddit, set basic auth
		req.SetBasicAuth(api.ClientID, api.clientSecret)
	}

	// set user agent
	req.Header.Set("User-Agent", api.UserAgent)

	return req, nil
}

// Get performs a GET request to the specified URL with the specified
// query parameters
// Don't forget to close the response body
func (api *RedditAPI) Get(u *url.URL, query url.Values) (*http.Response, error) {
	// add the GET query
	u.RawQuery = query.Encode()

	// create new request
	req, err := api.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}

	// log request
	if api.DebugMode {
		dump, err := httputil.DumpRequest(req, true)
		if err == nil {
			logrus.WithFields(logrus.Fields{
				"request_dump": string(dump),
			}).Info("sending request")
		}
	}

	// do request
	resp, err := api.Client.Do(req)
	if err != nil {
		return nil, err
	}

	// log response
	if api.DebugMode {
		dump, err := httputil.DumpResponse(resp, true)
		if err == nil {
			logrus.WithFields(logrus.Fields{
				"response_dump": string(dump),
			}).Info("response received")
		}
	}

	return resp, nil
}

// PostForm posts form data to the specified URL with the required
// authentication
// Don't forget to close the response body
func (api *RedditAPI) PostForm(u *url.URL, data url.Values) (*http.Response, error) {
	// create request body from data
	body := data.Encode()
	bodyReader := strings.NewReader(body)

	// create the request
	req, err := api.NewRequest(http.MethodPost, u, bodyReader)
	if err != nil {
		return nil, err
	}

	// set content type
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// log request
	if api.DebugMode {
		dump, err := httputil.DumpRequest(req, true)
		if err == nil {
			logrus.WithFields(logrus.Fields{
				"request_dump": string(dump),
			}).Info("sending request")
		}
	}

	// do request
	resp, err := api.Client.Do(req)
	if err != nil {
		return nil, err
	}

	// log response
	if api.DebugMode {
		dump, err := httputil.DumpResponse(resp, true)
		if err == nil {
			logrus.WithFields(logrus.Fields{
				"response_dump": string(dump),
			}).Info("response received")
		}
	}

	return resp, nil
}
