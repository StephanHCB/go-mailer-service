package email

import "github.com/gin-gonic/gin"

// Model for EmailDto.
//
// swagger:model emailDto
type EmailDto struct {
	// The email address to send to
	ToAddress string `json:"to_address"`
	// The email subject
	Subject   string `json:"subject"`
	// The email body
	Body      string `json:"body"`
}

// TODO this seems needed to express using a model

// Parameters for sending Emails
//
// swagger:parameters sendEmailEndpoint
type SendEmailParams struct {
	// in:body
	Body EmailDto
}

type EmailApi interface {
	// swagger:route GET /email/send email-tag sendEmailEndpoint
	// This will eventually send an email.
	//
	// responses:
	//   501: errorResponse
	SendEmail(*gin.Context)
}
