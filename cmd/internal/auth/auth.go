package auth

import (
	"errors"
	"net/http"
	"strings"
)

func GetBearerToken(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("no authorization header included")
	}

	token, ok := strings.CutPrefix(authHeader, "Bearer ")
	if !ok {
		return "", errors.New("authorization header must start with Bearer")
	}

	if token == "" {
		return "", errors.New("bearer token is empty")
	}

	return token, nil
}
