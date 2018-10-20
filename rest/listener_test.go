package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Phil192/rediq/storage"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"
)

var socket = "0.0.0.0:8082"

func init() {
}

func TestMain(m *testing.M) {
	var f io.Writer
	myCache := storage.NewCache(
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
	time.Sleep(3 * time.Second)
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

func TestGetBy(t *testing.T) {
	r := require.New(t)
	var innerArr = []string{"ok"}
	data := postItem{"testGetBy", innerArr, 5}
	j, err := json.Marshal(&data)
	r.NoError(err)
	resp, err := http.Post(
		fmt.Sprintf("http://%s/api/v1/set", socket),
		"application/json",
		bytes.NewBuffer(j),
	)
	r.NoError(err)
	r.Equal(200, resp.StatusCode)
	u, err := url.ParseRequestURI("http://" + socket)
	r.NoError(err)
	q := u.Query()
	q.Set("key", "testGetBy")
	q.Set("index", "0")
	u.RawQuery = q.Encode()
	u.Path = "/api/v1/getby/"
	r.NoError(err)
	r.Equal(200, resp.StatusCode)
	resp, err = http.Get(u.String())
	r.NoError(err)
	r.Equal(200, resp.StatusCode)
	bts, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	r.NoError(err)
	r.Contains(string(bts), "ok")
}

func TestSet(t *testing.T) {
	r := require.New(t)
	data := postItem{"testSet", "ok", 2}
	j, err := json.Marshal(&data)
	r.NoError(err)
	resp, err := http.Post(
		fmt.Sprintf("http://%s/api/v1/set", socket),
		"application/json",
		bytes.NewBuffer(j),
	)
	r.NoError(err)
	r.Equal(200, resp.StatusCode)
	resp, err = http.Get(fmt.Sprintf("http://%s/api/v1/get/testSet", socket))
	r.NoError(err)
	r.Equal(200, resp.StatusCode)
	bts, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	r.NoError(err)
	r.Contains(string(bts), "ok")
	time.Sleep(3 * time.Second)
	resp, err = http.Get(fmt.Sprintf("http://%s/api/v1/get/testSet", socket))
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
