package client

import (
	"time"
	"encoding/json"
	"net/http"
	"bytes"
	"fmt"
	"crypto/sha1"
	"io/ioutil"
)

type postItem struct {
	Key   string        `json:"key"`
	Value string        `json:"value"`
	TTL   time.Duration `json:"ttl"`
}

type User interface {
	Socket() string
	Post(string, string, string, time.Duration) ([]byte, error)
	Get(string, string) ([]byte, error)
	Delete(string, string) ([]byte, error)
}

type cacheClient struct {
	cli   *http.Client
	sock  string
	login string
	pass  string
}

func NewClient(sock, lgn, pass string) User {
	return &cacheClient{
		cli:   &http.Client{},
		sock:  sock,
		login: lgn,
		pass:  pass,
	}
}

func (c *cacheClient) Socket() string {
	return c.sock
}

func (c *cacheClient) Post(addr string, key, val string, dur time.Duration) ([]byte, error) {
	data := postItem{key, val, dur}
	j, err := json.Marshal(&data)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", addr, bytes.NewBuffer(j))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return c.sendRequest(req)
}

func (c *cacheClient) Get(addr, key string) ([]byte, error) {
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf("%s%s", addr, key),
		nil,
	)
	if err != nil {
		return nil, err
	}
	return c.sendRequest(req)
}

func (c *cacheClient) Delete(addr, key string) ([]byte, error) {
	req, err := http.NewRequest(
		"DELETE",
		fmt.Sprintf("%s%s", addr, key),
		nil,
	)
	if err != nil {
		return nil, err
	}
	return c.sendRequest(req)

}

func (c *cacheClient) sendRequest(req *http.Request) ([]byte, error) {
	hasher := sha1.New()
	_, err := hasher.Write([]byte(c.login + c.pass))
	req.Header.Set("token", fmt.Sprintf("%x", hasher.Sum(nil)))

	resp, err := c.cli.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}