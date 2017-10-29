// Copyright © 2016 Aaron Longwell
//
// Use of this source code is governed by an MIT licese.
// Details in the LICENSE file.

package trello

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
	"github.com/pkg/errors"
	"google.golang.org/appengine/urlfetch"
	"golang.org/x/net/context"
)

const DEFAULT_BASEURL = "https://api.trello.com/1"

type Client struct {
	client   *http.Client
	Logger   logger
	BaseURL  string
	Key      string
	Token    string
	throttle <-chan time.Time
	testMode bool
}

type logger interface {
	Debugf(string, ...interface{})
}

func NewClient(key, token string, ctx context.Context) *Client {
	return &Client{
		client:   urlfetch.Client(ctx),
		BaseURL:  DEFAULT_BASEURL,
		Key:      key,
		Token:    token,
		throttle: time.Tick(time.Second / 8), // Actually 10/second, but we're extra cautious
		testMode: false,
	}
}

func (c *Client) Throttle() {
	if !c.testMode {
		<-c.throttle
	}
}

func (c *Client) Get(path string, args Arguments, target interface{}) error {

	// Trello prohibits more than 10 seconds/second per token
	c.Throttle()

	params := args.ToURLValues()
	c.log("[trello] GET %s?%s", path, params.Encode())

	if c.Key != "" {
		params.Set("key", c.Key)
	}

	if c.Token != "" {
		params.Set("token", c.Token)
	}

	url := fmt.Sprintf("%s/%s", c.BaseURL, path)
	urlWithParams := fmt.Sprintf("%s?%s", url, params.Encode())

	req, err := http.NewRequest("GET", urlWithParams, nil)
	if err != nil {
		return errors.Wrapf(err, "Invalid GET request %s", url)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return errors.Wrapf(err, "HTTP request failure on %s", url)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return makeHttpClientError(url, resp)
	}

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(target)
	if err != nil {
		return errors.Wrapf(err, "JSON decode failed on %s", url)
	}

	return nil
}

func (c *Client) Put(path string, args Arguments, target interface{}) error {

	// Trello prohibits more than 10 seconds/second per token
	c.Throttle()

	params := args.ToURLValues()
	c.log("[trello] PUT %s?%s", path, params.Encode())

	if c.Key != "" {
		params.Set("key", c.Key)
	}

	if c.Token != "" {
		params.Set("token", c.Token)
	}

	url := fmt.Sprintf("%s/%s", c.BaseURL, path)
	urlWithParams := fmt.Sprintf("%s?%s", url, params.Encode())

	req, err := http.NewRequest("PUT", urlWithParams, nil)
	if err != nil {
		return errors.Wrapf(err, "Invalid PUT request %s", url)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return errors.Wrapf(err, "HTTP request failure on %s", url)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return makeHttpClientError(url, resp)
	}

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(target)
	if err != nil {
		return errors.Wrapf(err, "JSON decode failed on %s", url)
	}

	return nil
}

func (c *Client) Post(path string, args Arguments, target interface{}) error {

	// Trello prohibits more than 10 seconds/second per token
	c.Throttle()

	params := args.ToURLValues()
	c.log("[trello] POST %s?%s", path, params.Encode())

	if c.Key != "" {
		params.Set("key", c.Key)
	}

	if c.Token != "" {
		params.Set("token", c.Token)
	}

	url := fmt.Sprintf("%s/%s", c.BaseURL, path)
	urlWithParams := fmt.Sprintf("%s?%s", url, params.Encode())

	req, err := http.NewRequest("POST", urlWithParams, nil)
	if err != nil {
		return errors.Wrapf(err, "Invalid POST request %s", url)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.client.Do(req)
	if err != nil {
		return errors.Wrapf(err, "HTTP request failure on %s", url)
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrapf(err, "HTTP Read error on response for %s", url)
	}

	decoder := json.NewDecoder(bytes.NewBuffer(b))
	err = decoder.Decode(target)
	if err != nil {
		return errors.Wrapf(err, "JSON decode failed on %s:\n%s", url, string(b))
	}

	return nil
}

func (c *Client) log(format string, args ...interface{}) {
	if c.Logger != nil {
		c.Logger.Debugf(format, args)
	}
}
