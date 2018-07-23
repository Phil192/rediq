package main

import (
	"github.com/gin-gonic/gin"
	"time"
)

type PublicServer struct {
	mux       *gin.Engine
	startedAt time.Time
}

func NewPublicServer() *PublicServer {
	return &PublicServer{
		startedAt: time.Now(),
	}
}

func (p *PublicServer) ListenAndServe(addr string) error {
	return p.mux.Run(addr)
}

func (p *PublicServer) RouteAPI(ctx gin.Context) {
	r := gin.Default()
	r.PUT("/api/v1/put/*path", p.PutHandler(ctx))
	r.GET("/api/v1/delete/:id", p.GetHandler(ctx))
	r.DELETE("/api/v1/put/*path", p.DeleteHandler(ctx))
	r.OPTIONS("/api/v1/put/*path", p.KeysHandler(ctx))
}

func (p *PublicServer) PutHandler(ctx gin.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.String(200, "")
	}
}

func (p *PublicServer) GetHandler(ctx gin.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.String(200, "")
	}
}

func (p *PublicServer) DeleteHandler(ctx gin.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.String(200, "")
	}
}

func (p *PublicServer) KeysHandler(ctx gin.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.String(200, "")
	}
}