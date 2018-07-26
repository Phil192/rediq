package rest

import (
	"testing"
	"github.com/stretchr/testify/require"
	"net/http"
	"fmt"
	"os"
	"github.com/rediq/storage"
	"github.com/gin-gonic/gin"
	"log"
	"io"
	"bytes"
	"encoding/json"
	"io/ioutil"
	"time"
)

var socket = "0.0.0.0:8082"

func init() {
}

func TestMain(m *testing.M) {
	var f io.Writer
	myCache := storage.NewCache(
		storage.DefaultTTL(1),
		storage.DumpPath("./var/cache.dump"),
	)
	myCache.Run()
	app := NewApp(
		myCache,
		LogFile(f),
		SetSocket(socket),
	)
	app.RouteAPI(gin.Default())
	go func() {
		if err := app.ListenAndServe(); err != nil {
			log.Fatalf("listen: %s\n", err)
		}
	}()
	time.Sleep(3*time.Second)
	code := m.Run()
	myCache.Close()
	os.Exit(code)
}


func TestPing(t *testing.T) {
	r := require.New(t)
	resp, err := http.Get(fmt.Sprintf("http://%s/ping", socket))
	r.NoError(err)
	r.Equal(404, resp.StatusCode)
}

func TestGet(t *testing.T) {
	r := require.New(t)
	data := postItem{"testGet", "ok", 0}
	j, err := json.Marshal(&data)
	r.NoError(err)
	resp, err := http.Post(
		fmt.Sprintf("http://%s/api/v1/set", socket),
		"application/json",
		bytes.NewBuffer(j),
	)
	r.NoError(err)
	r.Equal(200, resp.StatusCode)
	resp, err = http.Get(fmt.Sprintf("http://%s/api/v1/get/testGet", socket))
	r.NoError(err)
	r.Equal(200, resp.StatusCode)
	bts, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	r.NoError(err)
	r.Contains(string(bts), "ok")
}

func TestSet(t *testing.T) {
	r := require.New(t)
	data := postItem{"testSet", "ok", 0}
	j, err := json.Marshal(&data)
	r.NoError(err)
	resp, err := http.Post(
		fmt.Sprintf("http://%s/api/v1/set", socket),
		"application/json",
		bytes.NewBuffer(j),
		)
	r.NoError(err)
	r.Equal(200, resp.StatusCode)
}

func TestSetWithTTL(t *testing.T) {
	r := require.New(t)
	data := postItem{"testSetTTL", "ok", 2}
	j, err := json.Marshal(&data)
	r.NoError(err)
	resp, err := http.Post(
		fmt.Sprintf("http://%s/api/v1/setWithTTL", socket),
		"application/json",
		bytes.NewBuffer(j),
	)
	r.NoError(err)
	r.Equal(200, resp.StatusCode)
	resp, err = http.Get(fmt.Sprintf("http://%s/api/v1/get/testSetTTL", socket))
	r.NoError(err)
	r.Equal(200, resp.StatusCode)
	bts, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	r.NoError(err)
	r.Contains( string(bts), "ok")
	time.Sleep(3*time.Second)
	resp, err = http.Get(fmt.Sprintf("http://%s/api/v1/get/testSetTTL", socket))
	r.NoError(err)
	r.Equal(404, resp.StatusCode)
}

func TestKeys(t *testing.T) {
	r := require.New(t)
	data := postItem{"testKeys", ".", 0}
	j, err := json.Marshal(&data)
	r.NoError(err)
	resp, err := http.Post(
		fmt.Sprintf("http://%s/api/v1/set", socket),
		"application/json",
		bytes.NewBuffer(j),
	)
	r.NoError(err)
	r.Equal(200, resp.StatusCode)
	resp, err = http.Get(fmt.Sprintf("http://%s/api/v1/keys/test*eys", socket))
	r.NoError(err)
	r.Equal(200, resp.StatusCode)
	bts, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	r.NoError(err)
	r.Contains(string(bts), "testKeys")
}

func TestRemove(t *testing.T) {
	r := require.New(t)
	data := postItem{"testRemove", "ok", 0}
	j, err := json.Marshal(&data)
	r.NoError(err)
	resp, err := http.Post(
		fmt.Sprintf("http://%s/api/v1/set", socket),
		"application/json",
		bytes.NewBuffer(j),
	)
	r.NoError(err)
	r.Equal(200, resp.StatusCode)
	req, err := http.NewRequest(
		"DELETE",
		fmt.Sprintf("http://%s/api/v1/remove/testRemove", socket),
		nil,
	)
	r.NoError(err)
	cli := http.Client{}
	resp, err = cli.Do(req)
	r.NoError(err)
	r.Equal(200, resp.StatusCode)
	resp, err = http.Get(fmt.Sprintf("http://%s/api/v1/get/testRemove", socket))
	r.NoError(err)
	r.Equal(404, resp.StatusCode)
}
