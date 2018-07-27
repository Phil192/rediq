package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/abiosoft/ishell"
	"github.com/fatih/color"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type User interface {
	Socket() string
	Post(addr string, data postItem) ([]byte, error)
	Get(addr, key string) ([]byte, error)
	Delete(addr, key string) ([]byte, error)
}

type cacheClient struct {
	cli   *http.Client
	sock  string
	login string
	pass  string
}

func newClient(sock, lgn, pass string) User {
	return &cacheClient{
		cli:   &http.Client{},
		sock:  sock,
		login: lgn,
		pass:  pass,
	}
}

type postItem struct {
	Key   string        `json:"key"`
	Value string        `json:"value"`
	TTL   time.Duration `json:"ttl"`
}

func main() {
	login := flag.String("login", "", "login for basic auth")
	password := flag.String("password", "", "password for basic auth")
	sock := flag.String("socket", "http://0.0.0.0:8081", "socket to request")
	flag.Parse()

	shell := ishell.New()
	cli := newClient(
		*sock,
		*login,
		*password,
	)

	resp, err := cli.Get(*sock, "/ping")
	if err != nil {
		fail("Can't connect to cache server", string(resp))
		return
	}
	notice("Rediq cache storage. Simple Interactive Client")
	shell.AddCmd(&ishell.Cmd{
		Name: "set",
		Help: "set value to cache with time to live",
		Func: func(c *ishell.Context) {
			if len(c.Args) >= 2 {
				fail("must be two or three values")
				return
			}
			u, err := url.ParseRequestURI(cli.Socket())
			if err != nil {
				fail(err)
				return
			}
			u.Path = "/api/v1/setWithTTL/"
			ttl, err := strconv.Atoi(c.Args[2])
			if err != nil {
				fail(err)
				return
			}
			data := postItem{c.Args[0], c.Args[1], time.Duration(ttl) * time.Second}
			body, err := cli.Post(u.String(), data)
			if err != nil {
				fail(err)
				return
			}
			success(string(body))
		},
	})
	shell.AddCmd(&ishell.Cmd{
		Name: "get",
		Help: "get value from cache",
		Func: func(c *ishell.Context) {
			u, err := url.ParseRequestURI(cli.Socket())
			if err != nil {
				fail(err)
				return
			}
			u.Path = "/api/v1/get/"
			body, err := cli.Get(u.String(), c.Args[0])
			if err != nil {
				fail(err)
				return
			}
			success(string(body))
		},
	})
	shell.AddCmd(&ishell.Cmd{
		Name: "getby",
		Help: "get value by index (int or string)",
		Func: func(c *ishell.Context) {
			if len(c.Args) != 2 {
				fail("must be two values")
				return
			}
			u, err := url.ParseRequestURI(cli.Socket())
			if err != nil {
				fail(err)
				return
			}
			q := u.Query()
			q.Set("key", c.Args[0])
			q.Set("index", c.Args[1])
			u.RawQuery = q.Encode()
			u.Path = "/api/v1/getby/"
			body, err := cli.Get(u.String(), "")
			if err != nil {
				fail(err)
				return
			}
			success(string(body))
		},
	})
	shell.AddCmd(&ishell.Cmd{
		Name: "remove",
		Help: "remove value from cache",
		Func: func(c *ishell.Context) {
			u, err := url.ParseRequestURI(cli.Socket())
			if err != nil {
				fail(err)
				return
			}
			u.Path = "/api/v1/remove/"
			body, err := cli.Delete(u.String(), c.Args[0])
			if err != nil {
				fail(err)
				return
			}
			success(string(body))
		},
	})
	shell.AddCmd(&ishell.Cmd{
		Name: "keys",
		Help: "get matching keys from cache",
		Func: func(c *ishell.Context) {
			u, err := url.ParseRequestURI(cli.Socket())
			if err != nil {
				fail(err)
				return
			}
			u.Path = "/api/v1/keys/"
			body, err := cli.Get(u.String(), c.Args[0])
			if err != nil {
				fail(err)
				return
			}
			c.Println(string(body))
		},
	})
	shell.Run()
}

func (c *cacheClient) Socket() string {
	return c.sock
}

func (c *cacheClient) Post(addr string, data postItem) ([]byte, error) {
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

func fail(a ...interface{}) {
	r := color.New(color.FgRed).SprintFunc()
	fmt.Println(r(a))
}

func success(a ...interface{}) {
	gr := color.New(color.FgGreen).SprintFunc()
	fmt.Println(gr(a))
}

func notice(a ...interface{}) {
	note := color.New(color.Bold, color.FgBlue).SprintlnFunc()
	fmt.Println(note(a))
}
