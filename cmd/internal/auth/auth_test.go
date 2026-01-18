package auth

import (
	"net/http"
	"testing"
)

func TestGetBearerToken(t *testing.T) {
	true_token := "jshdbjyrf328y4_3ubf187413ifub3iuf"
	h := http.Header{}
	h.Set("Authorization", "Bearer jshdbjyrf328y4_3ubf187413ifub3iuf")

	access_token, err := GetBearerToken(h)
	if err != nil {
		t.Fatalf("GetBearerToken Failer: %v", err)
	}

	if access_token == "" {
		t.Error("Expected non-empty token")
	}

	if access_token != true_token {
		t.Errorf("Expected token: %v, got token: %v", true_token, access_token)
	}
}

func TestGetBearerTokenSpace(t *testing.T) {
	true_token := "jshdbjyrf328y4 3ubf187413ifub3iuf"
	h := http.Header{}
	h.Set("Authorization", "Bearer jshdbjyrf328y4 3ubf187413ifub3iuf")

	access_token, err := GetBearerToken(h)
	if err != nil {
		t.Fatalf("GetBearerToken Failer: %v", err)
	}

	if access_token == "" {
		t.Error("Expected non-empty token")
	}

	if access_token != true_token {
		t.Errorf("Expected token: %v, got token: %v", true_token, access_token)
	}
}

func TestGetBearerTokenEmpty(t *testing.T) {
	h := http.Header{}
	h.Set("Authorization", "Bearer ")

	access_token, err := GetBearerToken(h)
	if err == nil {
		t.Fatalf("GetBearerToken Failed: %v", err)
	}

	if access_token != "" {
		t.Errorf("Expected token: %v, got token: %v", "", access_token)
	}
}
