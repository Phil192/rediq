package rest

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/rediq/storage"
	"io"
	"strconv"
	"os"
	"io/ioutil"
	"encoding/json"
)

type application struct {
	mux   *gin.Engine
	cache storage.Storer
}

func NewApp(c storage.Storer, logFile io.Writer) *application {
	gin.DefaultWriter = os.Stdout
	if logFile != nil {
		gin.DefaultWriter = io.MultiWriter(logFile)
	}
	app := &application{
		cache: c,
	}
	return app
}

func (a *application) ListenAndServe(addr string) error {
	return a.mux.Run(addr)
}

func (a *application) RouteAPI(r *gin.Engine) {
	r.POST("/api/v1/set", TokenAuthMiddleware(), a.SetHandler)
	r.POST("/api/v1/setWithTTL", TokenAuthMiddleware(), a.SetWithTTLHandler)
	r.GET("/api/v1/get/:key", TokenAuthMiddleware(), a.GetHandler)
	r.GET("/api/v1/content", TokenAuthMiddleware(), a.GetContentHandler)
	r.DELETE("/api/v1/remove/:key", TokenAuthMiddleware(), a.DeleteHandler)
	r.GET("/api/v1/keys/:key", TokenAuthMiddleware(), a.KeysHandler)
	a.mux = r

}
type postItem struct {
	Key string `json:"key"`
	Value string `json:"value"`
	TTL		int `json:"ttl"`
}

func (a *application) SetHandler(c *gin.Context) {
	var item postItem
	body := c.Request.Body

	data, err := ioutil.ReadAll(body)
	if err != nil {
		c.AbortWithStatus(500)
		return
	}
	if err := json.Unmarshal(data, &item); err != nil {
		c.AbortWithStatus(500)
		return
	}
	if item.Key == "" || item.Value == "" {
		c.AbortWithStatus(400)
		return
	}
	if err := a.cache.Set(item.Key, item.Value); err != nil {
		c.AbortWithStatus(500)
		return
	}
	c.String(200, fmt.Sprintf("%s/api/v1/get/%s", c.Request.Host, item.Key))
}

func (a *application) SetWithTTLHandler(c *gin.Context) {
	var item postItem
	body := c.Request.Body

	data, err := ioutil.ReadAll(body)
	if err != nil {
		c.AbortWithStatus(500)
		return
	}
	if err := json.Unmarshal(data, &item); err != nil {
		c.AbortWithStatus(500)
		return
	}
	if item.Key == "" || item.Value == "" {
		c.AbortWithStatus(400)
		return
	}
	if err := a.cache.SetWithTTL(item.Key, item.Value, item.TTL); err != nil {
		c.AbortWithStatus(500)
		return
	}
	c.String(200, c.Request.Host, fmt.Sprintf("/api/v1/get/%s", item.Key))

}

func (a *application) GetHandler(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.AbortWithStatus(400)
		return
	}
	val, err := a.cache.Get(key)
	if err == storage.ErrNotFound {
		c.AbortWithStatus(404)
		return
	} else if err != nil {
		c.AbortWithError(500, err)
		return
	}
	c.JSON(200, val)
}

func (a *application) GetContentHandler(c *gin.Context) {
	var resp []byte
	var err error

	key := c.Query("key")
	if key == "" {
		c.AbortWithStatus(400)
		return
	}
	content := c.Query("content")
	if content == "" {
		c.AbortWithStatus(400)
		return
	}
	index, err := strconv.Atoi(content)
	if err == nil {
		if index < 0 {
			c.AbortWithError(400, storage.ErrSubSeqType)
			return
		}
		resp, err = a.cache.GetContent(key, index)
	} else {
		resp, err = a.cache.GetContent(key, content)
	}
	if err == storage.ErrSubSeqType {
		c.AbortWithStatus(400)
		return
	} else if err != nil {
		c.AbortWithError(500, err)
		return
	}
	c.String(200, string(resp))

}

func (a *application) DeleteHandler(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.AbortWithStatus(400)
		return
	}
	if err := a.cache.Remove(key); err != nil {
		c.AbortWithError(500, err)
		return
	}
	c.Status(200)

}

func (a *application) KeysHandler(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.AbortWithStatus(400)
		return
	}
	matchings := a.cache.Keys(key)
	if len(*matchings) == 0 {
		c.AbortWithStatus(404)
		return
	}
	c.JSON(200, matchings)

}
