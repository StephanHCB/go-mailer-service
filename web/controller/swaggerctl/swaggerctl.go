package swaggerctl

import (
	"github.com/StephanHCB/go-autumn-web-swagger-ui"
	"github.com/gin-gonic/gin"
)

func SetupSwaggerRoutes(server *gin.Engine) {
	server.StaticFS("/swagger-ui", auwebswaggerui.Assets)
	server.StaticFile("swagger.json", "docs/swagger.json")
}