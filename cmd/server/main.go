package main

import (
	"github.com/gin-gonic/gin"
	"github.com/rediq/rest"
	"github.com/rediq/storage"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"flag"
)

func main() {
	logLevel := flag.Int("logLevel", 5, "set log level")
	//addr := flag.String("addr", ":8080", "http server address")
	//readTimeout := flag.Int("readTimeout", 10, "http read timeout")
	//writeTimeout := flag.Int("writeTimeout", 10, "http write timeout")
	//
	//defaultTtl := flag.Int("defaultTTL", 0, "default ttl in seconds for every entry")
	//nShards := flag.Int("shards", 1, "number of shards for concurrent writes")
	//
	//login := flag.String("login", "", "login for basic auth")
	//password := flag.String("password", "", "password for basic auth")
	//
	//filename := flag.String("file", "", "database path")
	//saveFreq := flag.Int("saveFreq", 500, "save to disk frequency in ms")
	//
	//logTo := flag.String("log", "var/cache.log", "Does not log if empty")

	flag.Parse()

	log.SetLevel(log.Level(*logLevel))
	f, err := os.OpenFile("./var/rediq.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0755)
	if err != nil {
		log.Warningln(err)
	}
	log.SetOutput(f)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	c := storage.NewCache(storage.DefaultTTL(8))
	go func() {
		for {
			select {
			case <-interrupt:
				c.Close()
				log.Fatalf("interrupted")
			}
		}
	}()
	a := rest.NewApp(c, f)
	a.RouteAPI(gin.Context{})
	if err := a.ListenAndServe("localhost:8081"); err != nil {
		log.Fatalf("listen: %s\n", err)

	}
}
