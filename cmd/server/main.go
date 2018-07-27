package main

import (
	"flag"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/rediq/rest"
	"github.com/rediq/storage"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
)

var ErrTokenNotFound = errors.New("token not found in os env")

func main() {
	var f io.Writer
	var err error

	logLevel := flag.Int("logLevel", 5, "set log level")
	sock := flag.String("socket", "0.0.0.0:8081", "socket to listen")
	shardsNum := flag.Uint("shards", 256, "max number of shards")
	itemsNum := flag.Uint("items", 2048, "max number of items in single shard")
	output := flag.Bool("stdout", false, "stdout or log")
	logTo := flag.String("log", "./var/cache.log", "log file")
	dump := flag.String("dump", "./var/cache.dump", "path to dump cache data")
	flag.Parse()

	log.SetLevel(log.Level(*logLevel))

	if *output {
		f, err = os.OpenFile(*logTo, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0755)
		if err != nil {
			log.Fatalln(err)
		}
		log.SetOutput(f)
	}
	c := storage.NewCache(
		storage.ShardsNum(*shardsNum),
		storage.ItemsPerShard(*itemsNum),
		storage.DumpPath(*dump),
	)
	c.Run()
	//if err := checkEnvToken(); err != nil {
	//	log.Fatalln(err)
	//}
	app := rest.NewApp(
		c,
		rest.LogFile(f),
		rest.SetSocket(*sock),
	)
	app.RouteAPI(gin.Default())
	if err := app.ListenAndServe(); err != nil {
		log.Fatalf("listen: %s\n", err)
	}
}

func checkEnvToken() error {
	token := os.Getenv("token")
	if token == "" {
		return ErrTokenNotFound
	}
	return nil
}
