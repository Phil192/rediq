package main

import (
	"strings"
	"github.com/abiosoft/ishell"
)

func main(){
	shell := ishell.New()
	shell.Println("Rediq. Simple Interactive Client")
	shell.AddCmd(&ishell.Cmd{
		Name: "set",
		Help: "set value",
		Func: func(c *ishell.Context) {

			c.Println("Hello", strings.Join(c.Args, " "))
		},
	})

	shell.Run()
}
