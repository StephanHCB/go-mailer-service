package email

import "github.com/gin-gonic/gin"

// --- models ---

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

// --- parameters and responses --- needed to use models

// Parameters for sending Emails
//
// swagger:parameters sendEmailParams
type SendEmailParams struct {
	// in:body
	Body EmailDto
}

// The send email response with just a success status
//
// swagger:response sendEmailResponse
type SendEmailResponse struct {
}

// --- routes ---

type EmailApi interface {
	// swagger:route POST /api/rest/v1/sendmail email-tag sendEmailParams
	// This will send an email.
	//
	// responses:
	//   200: sendEmailResponse
	//   400: errorResponse
	//   401: errorResponse
	//   403: errorResponse
	//   500: errorResponse
	SendEmail(*gin.Context)
}
