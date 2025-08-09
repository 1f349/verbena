package rest

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type Client struct {
	HttpClient   *http.Client
	host         *url.URL
	accessToken  string
	refreshToken string
}

func NewClient(host string, refreshToken string) (*Client, error) {
	parseUrl, err := url.Parse(host)
	if err != nil {
		return nil, err
	}
	return &Client{
		HttpClient:   http.DefaultClient,
		host:         parseUrl,
		accessToken:  "",
		refreshToken: refreshToken,
	}, nil
}

func doRequest(c *Client, method string, p string, r io.Reader) (*http.Response, error) {
	resp, err := doRequestInternal(c, method, p, r)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusUnauthorized {
		err2 := refreshToken(c)
		if err2 != nil {
			return nil, fmt.Errorf("failed to refresh token: %v: %w", err2, err)
		}

		resp, err = doRequestInternal(c, method, p, r)
		if err != nil {
			return nil, err
		}
	}

	return resp, nil
}

func refreshToken(c *Client) error {
	req, err := http.NewRequest(http.MethodPost, c.host.JoinPath("refresh-bot-token").String(), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.refreshToken)
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("refresh-bot-token returned unexpected status %v", resp.StatusCode)
	}
	var token struct {
		Token string `json:"token"`
	}
	err = json.NewDecoder(resp.Body).Decode(&token)
	if err != nil {
		return err
	}
	c.accessToken = token.Token
	return nil
}

func doRequestInternal(c *Client, method string, p string, r io.Reader) (*http.Response, error) {
	if c.accessToken == "" {
		// Return a fake unauthorized response to prevent sending an empty access token
		return &http.Response{StatusCode: http.StatusUnauthorized, Body: io.NopCloser(io.MultiReader())}, nil
	}
	u := c.host.JoinPath(p)
	req, err := http.NewRequest(method, u.String(), r)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	return c.HttpClient.Do(req)
}
