package healthctl

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func Create(server *gin.Engine) {
	SetupRoutes(server)
}

func SetupRoutes(server *gin.Engine) {
	server.GET("/health", Health)
}

// actual endpoint implementation

func Health(ctx *gin.Context) {
	ctx.Writer.WriteHeader(http.StatusOK)
}
