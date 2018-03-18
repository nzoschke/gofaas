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
	}

	tokenString := strings.TrimPrefix(header(e, "Authorization"), "Bearer ")

	key, err := base64.StdEncoding.DecodeString(os.Getenv("AUTH_HASH_KEY"))
	if err != nil {
		r.Body = fmt.Sprintf("{\"error\": %q}", err)
		r.StatusCode = 500
		return r, claims, errors.WithStack(err)
	}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return key, nil
	})
	if err != nil {
		r.Body = fmt.Sprintf("{\"error\": %q}", err)
		r.StatusCode = 401
		return r, claims, errors.WithStack(err)
	}

	if !token.Valid {
		return r, claims, errors.New("Invalid token")
	}

	r.StatusCode = 200
	return r, claims, nil
}

func header(e events.APIGatewayProxyRequest, name string) string {
	if v := e.Headers[name]; v != "" {
		return v
	}

	return e.Headers[strings.ToLower(name)]
}
