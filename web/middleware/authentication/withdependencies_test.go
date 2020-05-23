package authentication

// This also tests the code we pulled in via dependencies

import (
	"context"
	"github.com/go-http-utils/headers"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

const validtoken_demosecret_HS256_admin = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyLCJodHRwczovL2dpdGh1Yi5jb20vU3RlcGhhbkhDQi9nby1jYW1wYWlnbi1zZXJ2aWNlL3JvbGVzIjpbImFkbWluIl19.EZC_nxHsZKrNLK6BvFqJrgpqWMv8OnnjpxAwst3b9RA"

const validtoken_demosecret_HS256_noroles = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.6F1Mu1zkAGqlk65ndU8InIVa5N8LIhDuOQYr-V_x8Tk"

const validtoken_demosecret_HS256_rolesnolist = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyLCJodHRwczovL2dpdGh1Yi5jb20vU3RlcGhhbkhDQi9nby1jYW1wYWlnbi1zZXJ2aWNlL3JvbGVzIjoidGhpcyBpcyBub3QgYSBsaXN0In0.zBYIPGrt1jvhnESJjoaiqmy8fKj8GLZubFT2HAdIrU8"

const validtoken_demosecret_HS256_rolesnostring = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyLCJodHRwczovL2dpdGh1Yi5jb20vU3RlcGhhbkhDQi9nby1jYW1wYWlnbi1zZXJ2aWNlL3JvbGVzIjpbIm5vdGFsaXN0b2ZzdHJpbmdzIiw0ODgsImFkbWluIl19.B-GjMsTwK9DceZL-dzNke0Wb6g9UMS-bpFVIp5Lyo_c"

func tstPrepareContextFromMiddleware(t *testing.T, secret string, token string) (context.Context, error) {
	cut := createAndConfigureAuthenticationMiddleware(secret)

	r, err := http.NewRequest("GET", "/unimportant", nil)
	if err != nil {
		t.Fatal(err)
	}
	if token != "" {
		r.Header.Set(headers.Authorization, "Bearer "+token)
	}
	w := httptest.NewRecorder()

	err = cut.CheckJWT(w, r)
	return r.Context(), err
}

func TestMiddlewareExtraction_ValidAdminToken(t *testing.T) {
	ctx, err := tstPrepareContextFromMiddleware(t, "demosecret", validtoken_demosecret_HS256_admin)
	require.Nil(t, err)

	actualRawToken, err := ExtractRawTokenFromContext(ctx)
	require.Nil(t, err)
	require.Equal(t, validtoken_demosecret_HS256_admin, actualRawToken)

	err = CheckUserIsLoggedIn(ctx)
	require.Nil(t, err)

	err = CheckUserHasRole(ctx, "admin")
	require.Nil(t, err)

	err = CheckUserHasRole(ctx, "friend")
	require.NotNil(t, err)
	require.Equal(t, "user does not have required role 'friend'", err.Error())
}

func TestMiddlewareExtraction_NoToken(t *testing.T) {
	ctx, err := tstPrepareContextFromMiddleware(t, "demosecret", "")
	// no token is a valid situation for publicly available requests
	require.Nil(t, err)

	_, err = ExtractRawTokenFromContext(ctx)
	require.NotNil(t, err)
	require.Equal(t, "no token found in context", err.Error())

	err = CheckUserIsLoggedIn(ctx)
	require.NotNil(t, err)
	require.Equal(t, "no token found in context", err.Error())

	err = CheckUserHasRole(ctx, "admin")
	require.NotNil(t, err)
	require.Equal(t, "no token found in context", err.Error())
}

func TestMiddlewareExtraction_InvalidSignatureSecretMismatch(t *testing.T) {
	ctx, err := tstPrepareContextFromMiddleware(t, "some-other-secret", validtoken_demosecret_HS256_admin)
	require.NotNil(t, err)
	require.Equal(t, "Error parsing token: signature is invalid", err.Error())

	_, err = ExtractRawTokenFromContext(ctx)
	require.NotNil(t, err)
	require.Equal(t, "no token found in context", err.Error())

	err = CheckUserIsLoggedIn(ctx)
	require.NotNil(t, err)
	require.Equal(t, "no token found in context", err.Error())

	err = CheckUserHasRole(ctx, "admin")
	require.NotNil(t, err)
	require.Equal(t, "no token found in context", err.Error())
}

func TestMiddlewareExtraction_ValidTokenWithSyntaxError_NoRoles(t *testing.T) {
	ctx, err := tstPrepareContextFromMiddleware(t, "demosecret", validtoken_demosecret_HS256_noroles)
	require.Nil(t, err)

	err = CheckUserHasRole(ctx, "admin")
	require.NotNil(t, err)
	require.Equal(t, "claim key '" + RolesClaimKey + "' not found, or value not available", err.Error())
}

func TestMiddlewareExtraction_ValidTokenWithSyntaxError_RolesNotAList(t *testing.T) {
	ctx, err := tstPrepareContextFromMiddleware(t, "demosecret", validtoken_demosecret_HS256_rolesnolist)
	require.Nil(t, err)

	err = CheckUserHasRole(ctx, "admin")
	require.NotNil(t, err)
	require.Equal(t, "claim value for key '" + RolesClaimKey + "' was not a json list", err.Error())
}

func TestMiddlewareExtraction_ValidTokenWithSyntaxError_RolesNotAllStrings(t *testing.T) {
	ctx, err := tstPrepareContextFromMiddleware(t, "demosecret", validtoken_demosecret_HS256_rolesnostring)
	require.Nil(t, err)

	err = CheckUserHasRole(ctx, "admin")
	require.NotNil(t, err)
	require.Equal(t, "role name '488' was not a string", err.Error())
}
