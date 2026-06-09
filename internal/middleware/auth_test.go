package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"
)

// Issue #62: garante que tokens precisam ter iss em expectedIssuers e aud == expectedAudience.
func TestValidateJWT_IssAudEnforcement(t *testing.T) {
	const secret = "test-secret"
	expectedIssuers := []string{"auth.institutoitinerante.com.br", "pulse-service"}
	expectedAud := "pulse-service"

	cases := []struct {
		name   string
		claims map[string]interface{}
		want   bool
	}{
		{
			name:   "valid - aud string match",
			claims: map[string]interface{}{"exp": time.Now().Add(time.Hour).Unix(), "iss": "pulse-service", "aud": "pulse-service"},
			want:   true,
		},
		{
			name:   "valid - aud array contains expected",
			claims: map[string]interface{}{"exp": time.Now().Add(time.Hour).Unix(), "iss": "auth.institutoitinerante.com.br", "aud": []string{"pulse-service", "other"}},
			want:   true,
		},
		{
			name:   "reject - wrong aud (token de focus-hub)",
			claims: map[string]interface{}{"exp": time.Now().Add(time.Hour).Unix(), "iss": "auth.institutoitinerante.com.br", "aud": "focus-hub"},
			want:   false,
		},
		{
			name:   "reject - missing aud",
			claims: map[string]interface{}{"exp": time.Now().Add(time.Hour).Unix(), "iss": "pulse-service"},
			want:   false,
		},
		{
			name:   "reject - unknown issuer",
			claims: map[string]interface{}{"exp": time.Now().Add(time.Hour).Unix(), "iss": "evil.example.com", "aud": "pulse-service"},
			want:   false,
		},
		{
			name:   "reject - missing iss",
			claims: map[string]interface{}{"exp": time.Now().Add(time.Hour).Unix(), "aud": "pulse-service"},
			want:   false,
		},
		{
			name:   "reject - expired",
			claims: map[string]interface{}{"exp": time.Now().Add(-time.Hour).Unix(), "iss": "pulse-service", "aud": "pulse-service"},
			want:   false,
		},
		{
			name:   "reject - missing exp",
			claims: map[string]interface{}{"iss": "pulse-service", "aud": "pulse-service"},
			want:   false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			token := signToken(t, secret, tc.claims)
			got := validateJWT(token, secret, expectedIssuers, expectedAud)
			if got != tc.want {
				t.Fatalf("validateJWT = %v, want %v (claims=%v)", got, tc.want, tc.claims)
			}
		})
	}
}

func TestValidateJWT_BadSignatureRejected(t *testing.T) {
	const secret = "real-secret"
	claims := map[string]interface{}{
		"exp": time.Now().Add(time.Hour).Unix(),
		"iss": "pulse-service",
		"aud": "pulse-service",
	}
	token := signToken(t, "attacker-secret", claims)
	if validateJWT(token, secret, []string{"pulse-service"}, "pulse-service") {
		t.Fatal("expected token assinado com secret diferente a ser rejeitado")
	}
}

func signToken(t *testing.T, secret string, claims map[string]interface{}) string {
	t.Helper()
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	payloadBytes, err := json.Marshal(claims)
	if err != nil {
		t.Fatalf("marshal claims: %v", err)
	}
	payload := base64.RawURLEncoding.EncodeToString(payloadBytes)
	signingInput := header + "." + payload
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signingInput))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return signingInput + "." + sig
}
