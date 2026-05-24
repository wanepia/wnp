package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	base  string
	token string
	http  *http.Client
}

type APIError struct {
	Message string `json:"error"`
}

func New(base, token string) *Client {
	return &Client{
		base:  strings.TrimRight(base, "/"),
		token: token,
		http:  &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *Client) do(method, path string, body, out interface{}) error {
	var r io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		r = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, c.base+path, r)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		var urlErr *url.Error
		if errors.As(err, &urlErr) {
			return fmt.Errorf("cannot reach API at %s\n  check your URL: wnp config set-url <url>", c.base)
		}
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		var apiErr APIError
		msg := ""
		if json.Unmarshal(data, &apiErr) == nil && apiErr.Message != "" {
			msg = apiErr.Message
		} else {
			msg = fmt.Sprintf("HTTP %d", resp.StatusCode)
		}
		if resp.StatusCode == 401 {
			return fmt.Errorf("unauthorized: %s\n  check your token: wnp config set-token <token>", msg)
		}
		if resp.StatusCode == 403 {
			return fmt.Errorf("forbidden: %s\n  your token may not have admin privileges", msg)
		}
		if resp.StatusCode == 404 {
			return fmt.Errorf("not found: %s", msg)
		}
		return fmt.Errorf("%s", msg)
	}

	if out == nil || len(data) == 0 {
		return nil
	}
	return json.Unmarshal(data, out)
}

func (c *Client) Get(path string, out interface{}) error {
	return c.do("GET", path, nil, out)
}

func (c *Client) Post(path string, body, out interface{}) error {
	return c.do("POST", path, body, out)
}

func (c *Client) Put(path string, body, out interface{}) error {
	return c.do("PUT", path, body, out)
}

func (c *Client) Delete(path string) error {
	return c.do("DELETE", path, nil, nil)
}

func (c *Client) GetRaw(path string) (json.RawMessage, error) {
	var raw json.RawMessage
	err := c.do("GET", path, nil, &raw)
	return raw, err
}

func (c *Client) PostRaw(path string, body interface{}) (json.RawMessage, error) {
	var raw json.RawMessage
	err := c.do("POST", path, body, &raw)
	return raw, err
}

func (c *Client) PutRaw(path string, body interface{}) (json.RawMessage, error) {
	var raw json.RawMessage
	err := c.do("PUT", path, body, &raw)
	return raw, err
}
