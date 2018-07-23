package main

import (
	"fmt"
	"time"
)

func main() {
	//srv := NewPublicServer()
	//ctx := gin.Context{}
	//srv.RouteAPI(ctx)
	//srv.ListenAndServe("localhost:8070")
	c := NewCache()
	go c.Run()
	defer c.Close()

	c.Set("test", []byte(`{"key":"val"}`), 8)
	c.Set("tist", []byte("!!!!!"), 8)
	v := c.Keys("t*st")
	fmt.Printf("%v\n", v)
	time.Sleep(time.Second*15)

	fmt.Println("end")

}

