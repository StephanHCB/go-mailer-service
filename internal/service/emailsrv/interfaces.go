package emailsrv

import (
	"context"
	"github.com/StephanHCB/go-mailer-service/internal/entity"
)

type EmailService interface {
	NewInstance(ctx context.Context) *entity.Email

	SendEmail(ctx context.Context, email *entity.Email) error
}
