package emailctl

import (
	"encoding/json"
	"github.com/StephanHCB/go-mailer-service/api/v1/apierrors"
	"github.com/StephanHCB/go-mailer-service/api/v1/email"
	"github.com/StephanHCB/go-mailer-service/internal/service/emailsrv"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/thanhhh/gin-requestid"
	"net/http"
	"time"
)

type EmailController struct {
	s emailsrv.EmailService
}

func Create(server *gin.Engine, emailService emailsrv.EmailService) email.EmailApi {
	controller := &EmailController{s: emailService}
	controller.SetupRoutes(server)
	return controller
}

func (c *EmailController) SetupRoutes(server *gin.Engine) {
	server.POST("/api/rest/v1/sendmail", c.SendEmail)
}

func (c *EmailController) SendEmail(ginctx *gin.Context) {
	dto, err := parseBodyToEmailDto(ginctx)
	if err != nil {
		emailParseErrorHandler(ginctx, err)
		return
	}

	// technical validation would happen here

	ctx := ginctx.Request.Context()
	email := c.s.NewInstance(ctx)
	err = mapDtoToEmail(dto, email)
	if err != nil {
		emailParseErrorHandler(ginctx, err)
		return
	}

	err = c.s.SendEmail(ctx, email)
	if err != nil {
		emailSendErrorHandler(ginctx, err)
		return
	}
	ginctx.Writer.WriteHeader(http.StatusOK)
}

func parseBodyToEmailDto(ginctx *gin.Context) (*email.EmailDto, error) {
	decoder := json.NewDecoder(ginctx.Request.Body)
	dto := &email.EmailDto{}
	err := decoder.Decode(dto)
	if err != nil {
		dto = &email.EmailDto{}
	}
	return dto, err
}

func emailParseErrorHandler(ginctx *gin.Context, err error) {
	// TODO better way to deal with request related errors? Also how will it get requestId?
	ctx := ginctx.Request.Context()
	log.Ctx(ctx).Warn().Err(err).Msgf("email body could not be parsed: %v", err)
	errorHandler(ginctx, "email.parse.error", http.StatusBadRequest, []string{})
}

func emailSendErrorHandler(ginctx *gin.Context, err error) {
	ctx := ginctx.Request.Context()
	log.Ctx(ctx).Warn().Err(err).Msgf("error sending email: %v", err)
	errorHandler(ginctx, "email.send.error", http.StatusInternalServerError, []string{})
}

func errorHandler(ginctx *gin.Context, msg string, status int, details []string) {
	timestamp := time.Now().Format(time.RFC3339)
	requestId := requestid.GetReqID(ginctx)
	response := apierrors.ErrorDto{Message: msg, Timestamp: timestamp, Details: details, RequestId: requestId}
	ginctx.JSON(status, response)
}
