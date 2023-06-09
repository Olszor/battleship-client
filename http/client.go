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

func (c *Client) InitGame(coords []string, description string, nick string, targetNick string, wpbot bool) error {
	body := InitGameRequest{
		Coords:     coords,
		Desc:       description,
		Nick:       nick,
		TargetNick: targetNick,
		Wpbot:      wpbot,
	}
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

func (c *Client) Description() (*DescriptionResponse, error) {
	requestUrl, err := url.JoinPath(c.url, "/game/desc")
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

	var body DescriptionResponse
	err = json.Unmarshal(bodyJson, &body)
	if err != nil {
		return nil, fmt.Errorf("error deserializing body: %s", err)
	}

	return &body, nil
}

func (c *Client) Fire(coord string) (*FireResponse, error) {
	reqBody := FireRequest{
		Coord: coord,
	}
	bodyJson, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("error serializing FireRequest to json: %s", err)
	}

	requestUrl, err := url.JoinPath(c.url, "/game/fire")
	if err != nil {
		return nil, fmt.Errorf("error creating url: %s", err)
	}

	req, err := http.NewRequest(http.MethodPost, requestUrl, bytes.NewReader(bodyJson))
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

	bodyJson, err = io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading body: %s", err)
	}

	var resBody FireResponse
	err = json.Unmarshal(bodyJson, &resBody)
	if err != nil {
		return nil, fmt.Errorf("error deserializing body: %s", err)
	}

	return &resBody, nil
}

func (c *Client) List() (*[]ListResponse, error) {
	requestUrl, err := url.JoinPath(c.url, "/game/list")
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

	var resBody []ListResponse
	err = json.Unmarshal(bodyJson, &resBody)
	if err != nil {
		return nil, fmt.Errorf("error deserializing body: %s", err)
	}

	return &resBody, nil
}

func (c *Client) Refresh() error {
	requestUrl, err := url.JoinPath(c.url, "/game/refresh")
	if err != nil {
		return fmt.Errorf("error creating url: %s", err)
	}

	req, err := http.NewRequest(http.MethodGet, requestUrl, http.NoBody)
	if err != nil {
		return fmt.Errorf("error creating request: %s", err)
	}
	req.Header.Set("X-Auth-Token", c.token)

	res, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %s", err)
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	return nil
}

func (c *Client) Stats() (*StatsResponse, error) {
	requestUrl, err := url.JoinPath(c.url, "/stats")
	if err != nil {
		return nil, fmt.Errorf("error creating url: %s", err)
	}

	req, err := http.NewRequest(http.MethodGet, requestUrl, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %s", err)
	}

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

	var resBody StatsResponse
	err = json.Unmarshal(bodyJson, &resBody)
	if err != nil {
		return nil, fmt.Errorf("error deserializing body: %s", err)
	}

	return &resBody, nil
}

func (c *Client) PlayerStats(player string) (*PlayerStatsResponse, error) {
	requestUrl, err := url.JoinPath(c.url, "/stats/"+player)
	if err != nil {
		return nil, fmt.Errorf("error creating url: %s", err)
	}

	req, err := http.NewRequest(http.MethodGet, requestUrl, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %s", err)
	}

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

	var resBody PlayerStatsResponse
	err = json.Unmarshal(bodyJson, &resBody)
	if err != nil {
		return nil, fmt.Errorf("error deserializing body: %s", err)
	}

	return &resBody, nil
}

func (c *Client) Abandon() error {
	requestUrl, err := url.JoinPath(c.url, "/game/abandon")
	if err != nil {
		return fmt.Errorf("error creating url: %s", err)
	}

	req, err := http.NewRequest(http.MethodDelete, requestUrl, http.NoBody)
	if err != nil {
		return fmt.Errorf("error creating request: %s", err)
	}
	req.Header.Set("X-Auth-Token", c.token)

	res, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %s", err)
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	return nil
}
