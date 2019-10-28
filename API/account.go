package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"time"
)

// grant types
const (
	GrantTypePassword = "password"
)

// RedditAccount holds the data pertaining to a reddit account
type RedditAccount struct {
	API      *RedditAPI
	Username string
	// Password should not be stored

	Token *Token
}

// Token stores the authentication token and expiry time so that the
// validity of the token can be automatically verified before
// requests.
type Token struct {
	Token     string         `json:"access_token"`
	TokenType string         `json:"token_type"`
	Scope     string         `json:"scope"`
	ExpiresIn TokenExpiresIn `json:"expires_in"` // seconds
	Expiry    time.Time
	Error     string `json:"error"`
}

// TokenExpiresIn is a Duration with custom unmarshaller for
// unmarshalling the duration as seconds
type TokenExpiresIn time.Duration

// UnmarshalJSON decodes the JSON into this
func (t *TokenExpiresIn) UnmarshalJSON(data []byte) error {
	var int64_duration int64
	if err := json.Unmarshal(data, &int64_duration); err != nil {
		return err
	}
	fmt.Printf("%d\n", int64_duration)
	int64_duration *= int64(time.Second)
	*t = TokenExpiresIn(int64_duration)

	return nil
}

// PasswordLogin uses a password to authenticate, storing the access
// token in the RedditAccount. Returns an error.
func (a *RedditAccount) PasswordLogin(password string) error {
	// get the URL for logging in
	redditURL := GetRedditURL(RedditEndpointLogin)

	// construct POST data
	data := url.Values{
		"grant_type": {GrantTypePassword},
		"username":   {a.Username},
		"password":   {password},
	}

	// send request
	resp, err := a.API.PostForm(redditURL, data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// check response code
	if resp.StatusCode != 200 {
		return errors.New(fmt.Sprintf("bad status code: %d", resp.StatusCode))
	}

	// decode response into new token
	var token Token
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&token)
	if err != nil {
		return errors.New(fmt.Sprintf("unable to decode json: %s", err))
	}
	if token.Error != "" {
		return errors.New(fmt.Sprintf("reddit returned error: %s", token.Error))
	}
	if token.Token == "" {
		// JSON decoded but token is bad
		return errors.New(fmt.Sprintf("blank token"))
	}

	// calculate expiry time
	token.Expiry = time.Now().Add(time.Duration(token.ExpiresIn))

	// store token
	a.Token = &token

	return nil
}
