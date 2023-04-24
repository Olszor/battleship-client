package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	client http.Client
	url    string
	token  string
}

func NewClient(url string, timeout time.Duration) *Client {
	return &Client{
		client: http.Client{
			Timeout: timeout,
		},
		url: url,
	}
}

func (c *Client) InitGame() error {
	body := InitGameRequest{}
	bodyJson, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("error serializing InitGameRequestBody to json: %s", err)
	}

	requestUrl, err := url.JoinPath(c.url, "/game")
	if err != nil {
		return fmt.Errorf("error creating url: %s", err)
	}

	req, err := http.NewRequest(http.MethodPost, requestUrl, bytes.NewReader(bodyJson))
	if err != nil {
		return fmt.Errorf("error creating request: %s", err)
	}

	res, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %s", err)
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	token := res.Header.Get("X-Auth-Token")
	if token == "" {
		return fmt.Errorf("token is missing")
	}

	c.token = token

	return nil
}

func (c *Client) Board() ([]string, error) {
	requestUrl, err := url.JoinPath(c.url, "/game/board")
	if err != nil {
		return nil, fmt.Errorf("error creating url: %s", err)
	}

	req, err := http.NewRequest(http.MethodGet, requestUrl, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %s", err)
	}
	req.Header.Set("X-Auth-Token", c.token)

	res, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %s", err)
	}

	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	bodyJson, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading body: %s", err)
	}

	var body BoardResponse
	err = json.Unmarshal(bodyJson, &body)
	if err != nil {
		return nil, fmt.Errorf("error deserializing body: %s", err)
	}

	return body.Board, nil
}

func (c *Client) Status() (*StatusResponse, error) {
	requestUrl, err := url.JoinPath(c.url, "/game")
	if err != nil {
		return nil, fmt.Errorf("error creating url: %s", err)
	}

	req, err := http.NewRequest(http.MethodGet, requestUrl, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %s", err)
	}
	req.Header.Set("X-Auth-Token", c.token)

	res, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %s", err)
	}

	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	bodyJson, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading body: %s", err)
	}

	var body StatusResponse
	err = json.Unmarshal(bodyJson, &body)
	if err != nil {
		return nil, fmt.Errorf("error deserializing body: %s", err)
	}

	return &body, nil
}
