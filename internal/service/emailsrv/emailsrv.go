package emailsrv

import (
	"context"
	"github.com/StephanHCB/go-mailer-service/internal/entity"
	"github.com/armon/go-metrics"
	"github.com/rs/zerolog/log"
	"time"
)

type EmailServiceImpl struct {
	// TODO mail sending repository
	// TODO messaging repository
}

func Create() EmailService {
	service := &EmailServiceImpl{}
	return service
}

func (e *EmailServiceImpl) NewInstance(ctx context.Context) *entity.Email {
	return &entity.Email{}
}

func (e *EmailServiceImpl) SendEmail(ctx context.Context, email *entity.Email) error {
	err := validate(email)
	if err != nil {
		log.Ctx(ctx).Warn().Msgf("business validation for email failed - rejected: %v", err.Error())
		return err
	}

	defer metrics.MeasureSince([]string{"SendEmail"}, time.Now())

	// TODO implement call to send email

	// TODO implement messaging call

	return nil
}
