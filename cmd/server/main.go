package main

import (
	"github.com/rediq/rest"
	"github.com/rediq/storage"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"flag"
	"io"
	"github.com/gin-gonic/gin"
	"fmt"
	"crypto/sha1"
)

func main() {
	logLevel := flag.Int("logLevel", 5, "set log level")
	sock := flag.String("socket", "0.0.0.0:8081", "socket to listen")
	defaultTtl := flag.Int("defaultTTL", 8, "default ttl in seconds for every entry")
	shardsNum := flag.Uint("shards", 256, "number of shards")
	itemsNum := flag.Uint("items", 2048, "number of items in single shard")
	output := flag.Bool("stdout", false, "stdout or log")
	logTo := flag.String("log", "./var/cache.log", "log file")
	login := flag.String("login", "", "login for basic auth")
	pass := flag.String("password", "", "password for basic auth")
	flag.Parse()

	var f io.Writer
	var err error
	log.SetLevel(log.Level(*logLevel))

	if *output {
		f, err = os.OpenFile(*logTo, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0755)
		if err != nil {
			log.Warningln(err)
		}
		log.SetOutput(f)
	}
	c := storage.NewCache(
		storage.DefaultTTL(*defaultTtl),
		storage.ShardsNum(*shardsNum),
		storage.ItemsPerShard(*itemsNum),
		)
	go c.Run()
	aaa := make(map[string]string, 0)
	aaa["adad"] = "aaaaa"
	c.Set("test", aaa)

	fmt.Println(c.Keys("*"))
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	go func() {
		for {
			select {
			case <-interrupt:
				c.Close()
				log.Fatalf("interrupted")
			}
		}
	}()
	hasher := sha1.New()
	_, err = hasher.Write([]byte(*login + *pass))
	if err != nil {
		log.Fatalln(err)
	}
	os.Setenv("token", fmt.Sprintf("%x", hasher.Sum(nil)))
	a := rest.NewApp(c, f)
	a.RouteAPI(gin.Default())
	if err := a.ListenAndServe(*sock); err != nil {
		log.Fatalf("listen: %s\n", err)

	}
}
