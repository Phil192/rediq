package rest

import (
	"encoding/json"
	"fmt"
	"github.com/Phil192/rediq/storage"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

type application struct {
	mux   *gin.Engine
	cache storage.Storer
	opt   *listenerOptions
}

func NewApp(c storage.Storer, opts ...listenerOpt) *application {
	app := &application{
		cache: c,
		opt:   &listenerOptions{},
	}
	for _, o := range opts {
		if o != nil {
			o(app.opt)
		}
	}
	return app
}

func (a *application) ListenAndServe() error {
	return a.mux.Run(a.opt.socket)
}

func (a *application) RouteAPI(r *gin.Engine) {
	r.POST("/api/v1/set", TokenAuthMiddleware(), a.setHandler)
	r.GET("/api/v1/get/:key", TokenAuthMiddleware(), a.getHandler)
	r.DELETE("/api/v1/remove/:key", TokenAuthMiddleware(), a.deleteHandler)
	r.GET("/api/v1/keys/:key", TokenAuthMiddleware(), a.keysHandler)
	r.GET("/api/v1/getby/", TokenAuthMiddleware(), a.getByHandler)
	a.mux = r
}

type postItem struct {
	Key   string        `json:"key"`
	Value interface{}   `json:"value"`
	TTL   time.Duration `json:"ttl"`
}

func (a *application) setHandler(c *gin.Context) {
	var item postItem
	body := c.Request.Body

	data, err := ioutil.ReadAll(body)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	if err := json.Unmarshal(data, &item); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	if item.Key == "" || item.Value == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	if err := a.cache.Set(item.Key, item.Value, item.TTL); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.String(http.StatusOK, fmt.Sprintf("/api/v1/get/%s", item.Key))
}

func (a *application) getHandler(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	val, err := a.cache.Get(key)
	if err == storage.ErrNotFound {
		c.AbortWithError(http.StatusNotFound, err)
		return
	} else if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, val)
}

func (a *application) getByHandler(c *gin.Context) {
	var resp interface{}
	var err error

	key := c.Query("key")
	if key == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	index := c.Query("index")
	if index == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	indexInt, err := strconv.Atoi(index)
	if err == nil {
		if indexInt < 0 {
			c.AbortWithError(http.StatusBadRequest, storage.ErrSubSeqType)
			return
		}
		resp, err = a.cache.GetBy(key, indexInt)
	} else {
		resp, err = a.cache.GetBy(key, index)
	}
	if err == storage.ErrSubSeqType {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	} else if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (a *application) deleteHandler(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	if err := a.cache.Remove(key); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.Status(http.StatusOK)
}

func (a *application) keysHandler(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	matchings := a.cache.Keys(key)
	if len(matchings) == 0 {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	c.JSON(http.StatusOK, matchings)
}
