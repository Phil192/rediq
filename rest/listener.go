package rest

import (
	"github.com/gin-gonic/gin"
	"github.com/rediq/storage"
	"strconv"
	"time"
	"io"
	"fmt"
)

type application struct {
	mux       *gin.Engine
	startedAt time.Time
	cache     storage.Storer
}

func NewApp(c storage.Storer, logFile io.Writer) *application {
	gin.DefaultWriter = io.MultiWriter(logFile)

	app := &application{
		startedAt: time.Now(),
		cache:     c,
	}
	app.mux.Use(gin.Recovery())
	//app.mux.Use(TokenAuthMiddleware())
	return app
}

func (a *application) ListenAndServe(addr string) error {
	a.cache.Run()
	return a.mux.Run(addr)
}

func (a *application) RouteAPI(ctx gin.Context) {
	r := gin.Default()
	r.POST("/api/v1/set/", TokenAuthMiddleware(), a.SetHandler(ctx)) //TODO test me
	r.POST("/api/v1/setWithTTL/", a.SetWithTTLHandler(ctx))
	r.GET("/api/v1/get/:key", a.GetHandler(ctx))
	r.GET("/api/v1/content/:key", a.GetContentHandler(ctx))
	r.DELETE("/api/v1/remove/:key", a.DeleteHandler(ctx))
	r.GET("/api/v1/keys/:key", a.KeysHandler(ctx))
}

func (a *application) SetHandler(ctx gin.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.PostForm("key")
		value := c.PostForm("value")
		if key == "" || value == "" {
			c.AbortWithStatus(400)
		}
		if err := a.cache.Set(key, value); err != nil {
			c.AbortWithStatus(500)
		}
		c.String(200, c.Request.Host, fmt.Sprintf("/api/v1/get/%s", key))
	}
}

func (a *application) SetWithTTLHandler(ctx gin.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.PostForm("key")
		value := c.PostForm("value")
		ttl := c.PostForm("ttl")
		if key == "" || value == "" {
			c.AbortWithStatus(400)
		}
		if ttl == "" {
			ttl = "0"
		}
		num, err := strconv.Atoi(ttl)
		if err != nil {
			c.AbortWithStatus(500)
		}
		if num < 0 {
			c.AbortWithError(400, storage.ErrNegativeTTL)
		}
		if err := a.cache.SetWithTTL(key, value, uint64(num)); err != nil {
			c.AbortWithStatus(500)
		}
		c.String(200, c.Request.Host, fmt.Sprintf("/api/v1/get/%s", key))
	}
}

func (a *application) GetHandler(ctx gin.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.Param("key")
		if key == "" {
			c.AbortWithStatus(400)
		}
		val, err := a.cache.Get(key)
		if err == storage.ErrNotFound {
			c.AbortWithStatus(404)
		} else if err != nil {
			c.AbortWithError(500, err)
		}
		switch val.Type() {
		case storage.STR:
			data, ok := val.Body().(string)
			if !ok {
				break
			}
			c.String(200, data, val.TTL())
		case storage.ARRAY:
			c.JSON(200, val)
		case storage.MAPPING:
			c.JSON(200, val)
		}
		c.AbortWithStatus(500)
	}
}

func (a *application) GetContentHandler(ctx gin.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		var resp []byte
		var err error
		key := c.Param("key")
		if key == "" {
			c.AbortWithStatus(400)
		}
		content := c.Query("content")
		if content == "" {
			c.AbortWithStatus(400)
		}
		index, err := strconv.Atoi(content)
		if err == nil {
			if index < 0 {
				c.AbortWithError(400, storage.ErrSubSeqType)
			}
			resp, err = a.cache.GetContent(key, index)
		} else {
			resp, err = a.cache.GetContent(key, content)
		}
		if err == storage.ErrSubSeqType {
			c.AbortWithStatus(400)
		} else if err != nil {
			c.AbortWithError(500, err)
		}
		c.String(200, string(resp))
	}
}

func (a *application) DeleteHandler(ctx gin.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.Param("key")
		if key == "" {
			c.AbortWithStatus(400)
		}
		if err := a.cache.Remove(key); err != nil {
			c.AbortWithError(500, err)
		}
		c.Status(200)
	}
}

func (a *application) KeysHandler(ctx gin.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.Param("key")
		if key == "" {
			c.AbortWithStatus(400)
		}
		match := a.cache.Keys(key)
		c.JSON(200, match)
	}
}
