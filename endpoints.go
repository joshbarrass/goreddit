package reddit

import (
	"fmt"
	"net/url"
)

// schemes
const (
	HTTP  = "http"
	HTTPS = "https"
)

// constants for the reddit url
const (
	redditScheme = HTTPS
	redditHost   = "www.reddit.com"

	oauthScheme = HTTPS
	oauthHost   = "oauth.reddit.com"
)

// GetRedditURL returns the URL for a reddit endpoint
func GetRedditURL(endpoint string, formats ...interface{}) *url.URL {
	return getURL(fmt.Sprintf(endpoint, formats...), "reddit")
}

// GetOauthURL returns the URL for an Oauth endpoint
func GetOauthURL(endpoint string, formats ...interface{}) *url.URL {
	return getURL(fmt.Sprintf(endpoint, formats...), "oauth")
}

func getURL(endpoint, t string) *url.URL {
	var (
		scheme string
		host   string
	)

	switch t {
	case "reddit":
		scheme = redditScheme
		host = redditHost
	case "oauth":
		scheme = oauthScheme
		host = oauthHost
	default:
		return nil
	}

	return &url.URL{
		Scheme: scheme,
		Host:   host,
		Path:   endpoint,
	}
}

/* Reddit Endpoints */
const (
	RedditEndpointLogin = "/api/v1/access_token"
)

/* Oauth Endpoints */
const (
	OauthEndpointMe                 = "/api/v1/me"
	OauthEndpointStylesheet         = "/r/%s/stylesheet"
	OauthEndpointSetStylesheet      = "/r/%s/api/subreddit_stylesheet"
	OauthEndpointStylesheetTemplate = "/r/%s/about/stylesheet.json"
	OauthEndpointSubmitPost         = "/api/submit"
	OauthEndpointRequestSticky      = "/api/set_subreddit_sticky"
	OauthEndpointRequestContestMode = "/api/set_contest_mode"
	OauthEndpointRequestRemovePost  = "/api/remove"
	OauthEndpointComposeMessage     = "/api/compose"
)
