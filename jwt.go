package gofaas

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
)

// JWTClaims validates the token in the Authorization header
// It returns a response with standard headers and claims if valid
// And an error response and an error if invalid
func JWTClaims(e events.APIGatewayProxyRequest, claims jwt.Claims) (events.APIGatewayProxyResponse, jwt.Claims, error) {
	r := events.APIGatewayProxyResponse{
		Headers: map[string]string{
			"Access-Control-Allow-Origin": header(e, "Origin"),
		},
		StatusCode: 200,
	}

	// for convenience, "pass" auth if no hash key is set
	k := os.Getenv("AUTH_HASH_KEY")
	if k == "" {
		return r, claims, nil
	}

	key, err := base64.StdEncoding.DecodeString(k)
	if err != nil {
		r.Body = fmt.Sprintf("{\"error\": %q}", err)
		r.StatusCode = 500
		return r, claims, errors.WithStack(err)
	}

	tokenString := strings.TrimPrefix(header(e, "Authorization"), "Bearer ")
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return key, nil
	})
	if token != nil && !token.Valid {
		err = errors.New("Invalid token")
	}
	if err != nil {
		r.Body = fmt.Sprintf("{\"error\": %q}", err)
		r.StatusCode = 401
		return r, claims, errors.WithStack(err)
	}

	return r, claims, nil
}

func header(e events.APIGatewayProxyRequest, name string) string {
	if v := e.Headers[name]; v != "" {
		return v
	}

	return e.Headers[strings.ToLower(name)]
}
