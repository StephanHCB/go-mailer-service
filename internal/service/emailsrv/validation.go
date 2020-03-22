package emailsrv

import "github.com/StephanHCB/go-mailer-service/internal/entity"

func validate(email *entity.Email) error {
	// some business validation

	// example: email address must not be @mailinator.com
	return nil
}
