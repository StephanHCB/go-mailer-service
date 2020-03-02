package emailctl

import (
	"github.com/StephanHCB/go-mailer-service/api/v1/apierrors"
	"github.com/StephanHCB/go-mailer-service/api/v1/email"
	"github.com/gin-gonic/gin"
	"net/http"
)

type EmailController struct {
	// TODO ref to service instance
}

func Create(server *gin.Engine) email.EmailApi {
	controller := &EmailController{}
	controller.SetupRoutes(server)
	return controller
}

func (c *EmailController) SetupRoutes(server *gin.Engine) {
	server.GET("/email/send", c.SendEmail)
}

// actual endpoint implementation

func (c *EmailController) SendEmail(ctx *gin.Context) {
	response := apierrors.ErrorDto{Message: "Not Implemented Yet"}
	ctx.JSON(http.StatusNotImplemented, response)
}
