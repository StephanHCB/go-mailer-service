package authentication

import (
	"context"
	"fmt"
	"github.com/dgrijalva/jwt-go"
)

func ExtractRawTokenFromContext(ctx context.Context) (string, error) {
	token, ok := ctx.Value("user").(*jwt.Token)
	if !ok {
		return "", fmt.Errorf("no token found in context")
	}

	return token.Raw, nil
}
