package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// FloatTime implements a custom unmarshaller for decoding the float
// time into a time.Time-compatible object
type FloatTime time.Time

func (ft *FloatTime) UnmarshalJSON(data []byte) error {
	// ensure it's not a bool first -- reddit is dodgy like that
	_, err := strconv.ParseBool(string(data))
	if err == nil {
		// keep zero value
		return nil
	}
	ts, err := strconv.ParseFloat(string(data), 64)
	if err != nil {
		return errors.New(fmt.Sprintf("failed to parse float: %s", err))
	}
	int_ts := int64(ts)
	t := time.Unix(int_ts, 0)
	*ft = FloatTime(t)
	return nil
}

// BaseResponse stores "error" and "message" and provides a function
// for validating a response based on these
type BaseResponse struct {
	Err     int64  `json:"error"`
	Message string `json:"message"`
}

func (r *BaseResponse) Error() error {
	if r.Err != 0 {
		return errors.New(fmt.Sprintf("reddit error '%v': %s", r.Err, r.Message))
	}
	return nil
}

// BaseJSONResponse is for the JSON API and provides the necessary
// Error function for it
type BaseJSONResponse struct {
	JSON    map[string][][]string `json:"json"`
	Message string                `json:"message"` // indicates a different failure
}

func (r *BaseJSONResponse) Error() error {
	if r.Message != "" {
		return errors.New(fmt.Sprintf("reddit error: %s", r.Message))
	}
	errs, ok := r.JSON["errors"]
	if !ok {
		// no errors
		return nil
	}

	// check how many errors occurred
	if len(errs) == 0 {
		return nil
	}
	if len(errs) > 1 {
		return errors.New("SetStylesheet: too many errors")
	}

	// if only one error, construct a proper error
	errString := strings.Join(errs[0], " ")
	return errors.New(fmt.Sprintf("SetStylesheet: %s", errString))
}

// MeResponse is the response from a Me query
type MeResponse struct {
	BaseResponse

	Username         string    `json:"reddit_bot"`
	CommentKarma     int       `json:"comment_karma"`
	LinkKarma        int       `json:"link_karma"`
	Created          FloatTime `json:"created"`
	CreatedUTC       FloatTime `json:"created_utc"`
	HasMail          bool      `json:"has_mail"`
	HasModMail       bool      `json:"has_mod_mail"`
	HasVerifiedEmail bool      `json:"has_verified_email"`
	ID               string    `json:"id"`
	HasGold          bool      `json:"is_gold"`
	IsMod            bool      `json:"is_mod"`
	Over18           bool      `json:"over_18"`
}

type stylesheetTemplateIntermediaryResponse struct {
	BaseResponse

	Kind string                 `json:"kind"`
	Data StylesheetTemplateData `json:"data"`
}

// StylesheetTemplateData is the data returned in the stylesheet
// template request
type StylesheetTemplateData struct {
	Images     []StylesheetTemplateImage `json:"images"`
	Stylesheet string                    `json:"stylesheet"`
}

// StylesheetTemplateImage stores the data about an image that has
// been uploaded to a subreddit for use in the stylesheet
type StylesheetTemplateImage struct {
	URL  string `json:"url"`
	Link string `json:"link"`
	Name string `json:"name"`
}

type SetStylesheetResponse struct {
	BaseJSONResponse
}

type intermediateSubmitPostResponse struct {
	JSON submitPostJSON
}

func (r *intermediateSubmitPostResponse) Error() error {
	errs := r.JSON.Errors
	switch len(errs) {
	case 0:
		return nil
	case 1:
		return errors.New(fmt.Sprintf("SubmitPost: %s", strings.Join(errs[0], " ")))
	default:
		return errors.New("SubmitPost: too many errors")
	}
}

func (r *intermediateSubmitPostResponse) GetData() *SubmitPostData {
	return &r.JSON.Data
}

type submitPostJSON struct {
	Errors [][]string     `json:"errors"`
	Data   SubmitPostData `json:"data"`
}

type SubmitPostData struct {
	URL    string `json:"url"`
	Drafts int    `json:"drafts_count"`
	ID     string `json:"id"`
	Name   string `json:"name"`
}

type requestStickyResponse struct {
	BaseJSONResponse
}

type requestContestModeResponse struct {
	BaseJSONResponse
}

/* Post Handling */

// All of this is a bit of a shit-show -- it should probably get
// refactored at some point, but for now it works

type postListingIntermediary struct {
	Data postIntermediary1 `json:"data"`
}

type postIntermediary1 struct {
	Children []postIntermediary2 `json:"children"`
}

type postIntermediary2 struct {
	Data PostResponse `json:"data"`
}

type PostResponse struct {
	Subreddit     string    `json:"subreddit"`
	Saved         bool      `json:"saved"`
	GildCount     int       `json:"gilded"`
	Hidden        bool      `json:"hidden"`
	Downvotes     int64     `json:"downs"`
	Name          string    `json:"name"`
	ID            string    `json:"id"`
	Quarantined   bool      `json:"quarantine"`
	SubredditType string    `json:"subreddit_type"`
	Upvotes       int64     `json:"ups"`
	AuthorName    string    `json:"author_fullname"`
	CommentCount  int64     `json:"num_comments"`
	Score         int64     `json:"score"`
	Edited        FloatTime `json:"edited"`
	IsSelf        bool      `json:"is_self"`
	Archived      bool      `json:"archived"`
	NSFW          bool      `json:"over_18"`
	Removed       bool      `json:"removed"`
	Spoiler       bool      `json:"spoiler"`
	Locked        bool      `json:"locked"`
	SubredditName string    `json:"subreddit_id"`
	Author        string    `json:"author"`
	ContestMode   bool      `json:"contest_mode"`
	Approved      bool      `json:"approved"`
	Stickied      bool      `json:"stickied"`
	URL           string    `json:"url"`
	CreatedUTC    FloatTime `json:"created_utc"`
	Body          string    `json:"selftext"`
	Replies       []CommentResponse
}

type commentListingIntermediary struct {
	Data commentIntermediary1 `json:"data"`
}

type commentIntermediary1 struct {
	Children []commentIntermediary2 `json:"children"`
}

type commentIntermediary2 struct {
	Data CommentResponse `json:"data"`
}

type CommentResponse struct {
	Subreddit      string          `json:"subreddit"`
	Saved          bool            `json:"saved"`
	GildCount      int             `json:"gilded"`
	Downvotes      int64           `json:"downs"`
	Name           string          `json:"name"`
	SubredditType  string          `json:"subreddit_type"`
	Upvotes        int64           `json:"ups"`
	AuthorName     string          `json:"author_fullname"`
	Score          int64           `json:"score"`
	Edited         FloatTime       `json:"edited"`
	Archived       bool            `json:"archived"`
	Removed        bool            `json:"removed"`
	Spoiler        bool            `json:"spoiler"`
	Locked         bool            `json:"locked"`
	SubredditName  string          `json:"subreddit_id"`
	Author         string          `json:"author"`
	ContestMode    bool            `json:"contest_mode"`
	Approved       bool            `json:"approved"`
	Stickied       bool            `json:"stickied"`
	CreatedUTC     FloatTime       `json:"created_utc"`
	Body           string          `json:"body"`
	ParentID       string          `json:"parent_id"`
	RepliesListing json.RawMessage `json:"replies"`
	Replies        []*CommentResponse
}

func (parentComment *CommentResponse) DecodeReplies() error {
	// empty the RepliesListing when done to free up some memory
	// if nothing went wrong
	failed := true
	defer func() {
		if !failed {
			parentComment.RepliesListing = make([]byte, 0)
		}
		return
	}()

	if string(parentComment.RepliesListing) == `""` {
		// no replies to decode
		failed = false
		return nil
	}
	// unmarshal the replies listing into a comment listing
	var commentListing commentListingIntermediary
	err := json.Unmarshal(parentComment.RepliesListing, &commentListing)
	if err != nil {
		return err
	}

	for _, comment := range commentListing.Data.Children {
		parentComment.Replies = append(parentComment.Replies, &comment.Data)
		err = comment.Data.DecodeReplies()
		if err != nil {
			return err
		}
	}

	failed = false
	return nil
}

// CommentsByScore implements a sort.Interface for sorting comments by
// score, from lowest to highest
type CommentsByScore []CommentResponse

func (a CommentsByScore) Len() int           { return len(a) }
func (a CommentsByScore) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a CommentsByScore) Less(i, j int) bool { return a[i].Score < a[j].Score }

// CommentsByScoreDescending implements a sort.Interface for sorting
// comments by score, from highest to lowest
type CommentsByScoreDescending CommentsByScore

func (a CommentsByScoreDescending) Len() int           { return len(a) }
func (a CommentsByScoreDescending) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a CommentsByScoreDescending) Less(i, j int) bool { return a[i].Score > a[j].Score }

type requestRemovePostResponse struct {
	BaseResponse
}

// ComposeMessageResponse is the response from composing a message
type ComposeMessageResponse struct {
	BaseJSONResponse
}
