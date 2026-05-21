package jwtx_test

import (
	"testing"
	"mall-server/pkg/jwtx"
)

func TestGenerateAndParseAdminToken(t *testing.T) {
	token, err := jwtx.GenerateAdminToken(1, "admin")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	claims, err := jwtx.ParseAdminToken(token)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if claims.AdminID != 1 || claims.Username != "admin" || !claims.IsAdmin {
		t.Fatalf("wrong claims: %+v", claims)
	}
}

func TestParseAdminToken_RejectUserToken(t *testing.T) {
	userToken, _ := jwtx.GenerateToken(1, "user")
	_, err := jwtx.ParseAdminToken(userToken)
	if err == nil {
		t.Fatal("expected error for user token")
	}
}
