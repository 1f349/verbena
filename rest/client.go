package rest

import (
	"io"
	"net/http"
	"net/url"
)

type Client struct {
	HttpClient *http.Client
	Host       string
}

func doRequest(c *Client, method string, p string, r io.Reader) (*http.Response, error) {
	u := &url.URL{
		Scheme: "https",
		Host:   c.Host,
		Path:   p,
	}
	req, err := http.NewRequest(method, u.String(), r)
	if err != nil {
		return nil, err
	}
	return c.HttpClient.Do(req)
}
