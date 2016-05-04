package atom

import (
	"net/url"
)

type Config struct {
	// ironSource.atom API key.
	ApiKey string

	// byte-array representation of the ApiKey. used to hmac requests.
	apikey []byte

	// Url is the url to send the request to. default to "https://track.atom-data.io"
	Url *url.URL

	// Client is the Poster interface implementation.
	Client Poster
}

// defaults for configuration
func (c *Config) defaults() {
	if len(c.ApiKey) == 0 {
		panic("atom: Invalid api key")
	}
	c.apikey = []byte(c.ApiKey)
	if c.Url == nil {
		c.Url, _ = url.Parse("https://track.atom-data.io")
	}
	if c.Client == nil {
		c.Client = &client{c.Url}
	}
}
