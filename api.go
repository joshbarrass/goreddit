package reddit

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
)

// decodeJSON takes a body and a pointer to decode data into
func decodeJSON(body io.Reader, p interface{}) error {
	d := json.NewDecoder(body)
	err := d.Decode(p)
	return err
}

// RequestMe queries the "me" API endpoint
func (api *RedditAPI) RequestMe() (*MeResponse, error) {
	url := GetOauthURL(OauthEndpointMe)
	resp, err := api.Get(url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response MeResponse
	if err := decodeJSON(resp.Body, &response); err != nil {
		return nil, err
	}
	if err := response.Error(); err != nil {
		return nil, err
	}

	return &response, nil
}

// RequestStylesheet gets the stylesheet of a particular subreddit
func (api *RedditAPI) RequestStylesheet(subreddit string) (string, error) {
	url := GetOauthURL(OauthEndpointStylesheet, subreddit)
	resp, err := api.Get(url, nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// check for 200
	if resp.StatusCode != http.StatusOK {
		return "", errors.New(fmt.Sprintf("returned status code %d", resp.StatusCode))
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	body := string(bodyBytes)

	return body, nil
}

// RequestStylesheetTemplate gets the stylesheet template (with
// e.g. %% %% for images instead of actual urls) for a particular
// subrededit
func (api *RedditAPI) RequestStylesheetTemplate(subreddit string) (*StylesheetTemplateData, error) {
	u := GetOauthURL(OauthEndpointStylesheetTemplate, subreddit)

	// send request
	resp, err := api.Get(u, url.Values{
		"raw_json": {"1"},
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// decode response
	var response stylesheetTemplateIntermediaryResponse
	if err := decodeJSON(resp.Body, &response); err != nil {
		return nil, err
	}
	if err := response.Error(); err != nil {
		return nil, err
	}

	// verify that the kind is as expected
	if response.Kind != "stylesheet" {
		return nil, errors.New(fmt.Sprintf("unexpected kind: %s", response.Kind))
	}

	// return only the data -- the wrapper is useless
	return &response.Data, nil
}

// RequestSetStylesheet sets the stylesheet for a subreddit
func (api *RedditAPI) RequestSetStylesheet(subreddit, stylesheet, reason string) (*SetStylesheetResponse, error) {
	u := GetOauthURL(OauthEndpointSetStylesheet, subreddit)

	// construct post data
	data := url.Values{
		"api_type":            {"json"},
		"op":                  {"save"},
		"reason":              {reason},
		"stylesheet_contents": {stylesheet},
	}

	// send request
	resp, err := api.PostForm(u, data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response SetStylesheetResponse
	if err := decodeJSON(resp.Body, &response); err != nil {
		return nil, err
	}
	if err := response.Error(); err != nil {
		return nil, err
	}

	// bodyBytes, err := ioutil.ReadAll(resp.Body)
	// if err != nil {
	// 	return nil, err
	// }
	// body := string(bodyBytes)
	// logrus.Info(body)

	return &response, nil
}

func (api *RedditAPI) RequestSubmitTextPost(subreddit, title, text string, ad, nsfw, spoiler, sendReplies bool) (*SubmitPostData, error) {
	// TODO: initial data validation

	u := GetOauthURL(OauthEndpointSubmitPost)

	// construct post data
	data := url.Values{
		"api_type":    {"json"},
		"kind":        {"self"},
		"nsfw":        {"false"},
		"resubmit":    {"false"},
		"sendreplies": {"false"},
		"spoiler":     {"false"},
		"sr":          {subreddit},
		"text":        {text},
		"title":       {title},
	}

	// send request
	resp, err := api.PostForm(u, data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response intermediateSubmitPostResponse
	if err := decodeJSON(resp.Body, &response); err != nil {
		return nil, err
	}
	if err := response.Error(); err != nil {
		return nil, err
	}
	postData := response.GetData()
	if postData.ID == "" {
		return nil, errors.New("empty post ID")
	}

	// bodyBytes, err := ioutil.ReadAll(resp.Body)
	// if err != nil {
	// 	return nil, err
	// }
	// body := string(bodyBytes)
	// logrus.Info(body)

	return postData, nil
}

// RequestSticky allows setting a post to sticky
// set num to -1 for bottom
func (api *RedditAPI) RequestSticky(subreddit string, name string, state bool, num int) error {
	u := GetOauthURL(OauthEndpointRequestSticky)

	// construct post data
	data := url.Values{
		"api_type": {"json"},
		"id":       {name},
		"r":        {subreddit},
		//"to_profile": {"false"},
	}
	if state {
		// making post sticky
		data["state"] = []string{"true"}
		if num >= 0 && num <= 4 {
			data["num"] = []string{strconv.Itoa(num)}
		}
	} else {
		// unsticky
		data["state"] = []string{"false"}
	}

	// send request
	resp, err := api.PostForm(u, data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var response requestStickyResponse
	if err := decodeJSON(resp.Body, &response); err != nil {
		return err
	}
	if err := response.Error(); err != nil {
		return err
	}

	// bodyBytes, err := ioutil.ReadAll(resp.Body)
	// if err != nil {
	// 	return err
	// }
	// body := string(bodyBytes)
	// logrus.Info(body)

	return nil
}

// RequestContestMode allows setting a post to contest mode
func (api *RedditAPI) RequestContestMode(name string, state bool) error {
	u := GetOauthURL(OauthEndpointRequestContestMode)

	// construct post data
	data := url.Values{
		"api_type": {"json"},
		"id":       {name},
	}
	if state {
		// enable
		data["state"] = []string{"true"}
	} else {
		// disable
		data["state"] = []string{"false"}
	}

	// send request
	resp, err := api.PostForm(u, data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var response requestContestModeResponse
	if err := decodeJSON(resp.Body, &response); err != nil {
		return err
	}
	if err := response.Error(); err != nil {
		return err
	}

	// bodyBytes, err := ioutil.ReadAll(resp.Body)
	// if err != nil {
	// 	return err
	// }
	// body := string(bodyBytes)
	// logrus.Info(body)

	return nil
}

// RequestPostJSON gets the JSON for a particular post
func (api *RedditAPI) RequestPostJSON(u *url.URL) (*PostResponse, error) {
	path := u.Path
	// remove trailing slash
	for path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}

	// add .json
	path = path + ".json"

	// change to oauth
	u.Host = oauthHost

	// reconstruct
	u.Path = path

	// get json
	resp, err := api.Get(u, url.Values{
		"raw_json": {"1"},
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// decode into arrays
	var arrays []json.RawMessage
	err = decodeJSON(resp.Body, &arrays)
	if err != nil {
		return nil, err
	}

	// decode the first item into a post
	var posts postListingIntermediary
	err = json.Unmarshal(arrays[0], &posts)
	if err != nil {
		return nil, err
	}
	post := posts.Data.Children[0].Data

	// decode the second item into a comments listing
	var commentsListing commentListingIntermediary
	err = json.Unmarshal(arrays[1], &commentsListing)
	if err != nil {
		return nil, err
	}

	// restructure the comments listing into an array of comments
	var comments = []CommentResponse{}
	for _, comment := range commentsListing.Data.Children {
		err = comment.Data.DecodeReplies()
		if err != nil {
			return nil, err
		}
		comments = append(comments, comment.Data)
	}

	// store the comments in the post
	post.Replies = comments

	return &post, nil
}

// RequestRemovePost removes a post as a moderator. Spam specifies
// whether or not to remove it as spam
func (api *RedditAPI) RequestRemovePost(name string, spam bool) error {
	u := GetOauthURL(OauthEndpointRequestRemovePost)

	// construct post data
	data := url.Values{
		"id": {name},
	}
	if spam {
		data["spam"] = []string{"true"}
	} else {
		data["spam"] = []string{"false"}
	}

	// send request
	resp, err := api.PostForm(u, data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var response requestRemovePostResponse
	if err := decodeJSON(resp.Body, &response); err != nil {
		return err
	}
	if err := response.Error(); err != nil {
		return err
	}

	// bodyBytes, err := ioutil.ReadAll(resp.Body)
	// if err != nil {
	// 	return err
	// }
	// body := string(bodyBytes)
	// logrus.Info(body)

	return nil
}

// ComposeMessage sends a message to another user
func (api *RedditAPI) ComposeMessage(to, subject, text string) error {
	u := GetOauthURL(OauthEndpointComposeMessage)

	// construct post data
	data := url.Values{
		"api_type": {"json"},
		"to":       {to},
		"subject":  {subject},
		"text":     {text},
	}

	// send request
	resp, err := api.PostForm(u, data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var response requestRemovePostResponse
	if err := decodeJSON(resp.Body, &response); err != nil {
		return err
	}
	if err := response.Error(); err != nil {
		return err
	}

	// bodyBytes, err := ioutil.ReadAll(resp.Body)
	// if err != nil {
	// 	return err
	// }
	// body := string(bodyBytes)
	// logrus.Info(body)

	return nil
}
