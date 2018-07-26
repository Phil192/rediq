package rest

import (
	"github.com/gin-gonic/gin"
	"io"
	"os"
)

type listenerOpt func(o *listenerOptions)

type listenerOptions struct {
	socket string
	engine *gin.Engine
}

func LogFile(logFile io.Writer) listenerOpt {
	return func(o *listenerOptions) {
		gin.DefaultWriter = os.Stdout
		if logFile != nil {
			gin.DefaultWriter = io.MultiWriter(logFile)
		}
	}
}

func SetSocket(sock string) listenerOpt {
	return func(o *listenerOptions) {
		o.socket = sock
	}
}
