package main

import (
	"flag"
	"fmt"
	"github.com/abiosoft/ishell"
	"github.com/fatih/color"
	"github.com/rediq/client"
	"net/url"
	"strconv"
	"time"
)

func main() {
	login := flag.String("login", "", "login for basic auth")
	password := flag.String("password", "", "password for basic auth")
	sock := flag.String("socket", "http://0.0.0.0:8081", "socket to request")
	flag.Parse()

	shell := ishell.New()
	cli := client.NewClient(
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
			if len(c.Args) < 3 {
				fail("must be three values")
				return
			}
			u, err := url.ParseRequestURI(cli.Socket())
			if err != nil {
				fail(err)
				return
			}
			u.Path = "/api/v1/set/"
			ttl, err := strconv.Atoi(c.Args[2])
			if err != nil {
				fail(err)
				return
			}
			dur := time.Duration(ttl) * time.Second
			body, err := cli.Post(u.String(), c.Args[0], c.Args[1], dur)
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
