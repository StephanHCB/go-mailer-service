package emailctl

import (
	"github.com/StephanHCB/go-mailer-service/api/v1/email"
	"github.com/StephanHCB/go-mailer-service/internal/entity"
)

func mapDtoToEmail(dto *email.EmailDto, c *entity.Email) error {
	c.Subject = dto.Subject
	c.Body	= dto.Body
	c.ToAddress = dto.ToAddress
	return nil
}
