package authentication

import (
	"context"
	"fmt"
	"github.com/dgrijalva/jwt-go"
)

// TODO role namespace url should come from config
const RolesClaimKey = "https://github.com/StephanHCB/go-campaign-service/roles"

func extractClaimFromTokenInContext(ctx context.Context, claimKey string) (interface{}, error) {
	token, ok := ctx.Value("user").(*jwt.Token)
	if !ok {
		return "", fmt.Errorf("no token found in context")
	}

	claimValue, ok := token.Claims.(jwt.MapClaims)[claimKey]
	if !ok {
		return "", fmt.Errorf("claim key '%s' not found, or value not available", claimKey)
	}

	return claimValue, nil
}

func CheckUserIsLoggedIn(ctx context.Context) error {
	_, err := extractClaimFromTokenInContext(ctx, "sub")
	return err
}

func CheckUserHasRole(ctx context.Context, role string) error {
	roles, err := extractClaimFromTokenInContext(ctx, RolesClaimKey)
	if err != nil {
		return err
	}

	rolesList, ok := roles.([]interface{})
	if !ok {
		return fmt.Errorf("claim value for key '%s' was not a json list", RolesClaimKey)
	}

	for _, roleValue := range rolesList {
		roleValueString, ok := roleValue.(string)
		if !ok {
			return fmt.Errorf("role name '%v' was not a string", roleValue)
		}
		if roleValueString == role {
			return nil
		}
	}

	return fmt.Errorf("user does not have required role '%s'", role)
}
