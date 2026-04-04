package utils

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestJWT_RoundTrip(t *testing.T) {
	j := NewJWTManager("access-secret-key-for-tests-32b!", "refresh-secret-key-for-tests-32b")
	uid := uuid.MustParse("01940000-0000-7000-8000-000000000099")

	at, err := j.GenerateAccessToken(uid, "user")
	if err != nil {
		t.Fatal(err)
	}
	ac, err := j.VerifyAccessToken(at)
	if err != nil {
		t.Fatal(err)
	}
	if ac.UserID != uid.String() || ac.Role != "user" {
		t.Fatalf("access claims %+v", ac)
	}

	rt, err := j.GenerateRefreshToken(uid, "admin")
	if err != nil {
		t.Fatal(err)
	}
	rc, err := j.VerifyRefreshToken(rt)
	if err != nil {
		t.Fatal(err)
	}
	if rc.Role != "admin" {
		t.Fatalf("refresh %+v", rc)
	}
}

func TestJWT_VerifyAccessToken_WrongSecret(t *testing.T) {
	j1 := NewJWTManager("secret-one-32bytes-minimum-length!!", "refresh-one-32bytes-minimum-length!")
	j2 := NewJWTManager("secret-two-32bytes-minimum-length!!", "refresh-two-32bytes-minimum-length!")
	tok, err := j1.GenerateAccessToken(uuid.New(), "user")
	if err != nil {
		t.Fatal(err)
	}
	_, err = j2.VerifyAccessToken(tok)
	if err == nil {
		t.Fatal("expected verify failure")
	}
}

func TestJWT_TTLConfigured(t *testing.T) {
	j := NewJWTManager("a", "b")
	if j.AccessTokenTTL != 15*time.Minute || j.RefreshTokenTTL != 7*24*time.Hour {
		t.Fatalf("ttl access %v refresh %v", j.AccessTokenTTL, j.RefreshTokenTTL)
	}
}
