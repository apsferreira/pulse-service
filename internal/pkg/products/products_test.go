package products

import (
	"os"
	"testing"
)

func TestNormalize(t *testing.T) {
	cases := map[string]string{
		"Nitro":     "nitro",
		"  NITRO  ": "nitro",
		"nitro":     "nitro",
		"":          "",
		"  ":        "",
		"SocialMake": "socialmake",
	}
	for in, want := range cases {
		if got := Normalize(in); got != want {
			t.Errorf("Normalize(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestIsKnown_Defaults(t *testing.T) {
	ResetForTest()
	os.Unsetenv("PULSE_KNOWN_PRODUCTS")

	for _, p := range []string{"libri", "nitro", "brio", "socialmake", "storemake", "pulse", "default"} {
		if !IsKnown(p) {
			t.Errorf("expected %q to be known by default", p)
		}
	}
	if IsKnown("Gemini") {
		t.Errorf("unexpected: capitalized product should not match (caller must Normalize first)")
	}
	if IsKnown("desconhecido") {
		t.Errorf("unexpected: 'desconhecido' should not be known")
	}
}

func TestIsKnown_EnvOverride(t *testing.T) {
	t.Setenv("PULSE_KNOWN_PRODUCTS", "alpha, Beta , gamma")
	ResetForTest()
	defer ResetForTest()

	if !IsKnown("alpha") || !IsKnown("beta") || !IsKnown("gamma") {
		t.Errorf("env-override allowlist should normalize entries: alpha=%v beta=%v gamma=%v",
			IsKnown("alpha"), IsKnown("beta"), IsKnown("gamma"))
	}
	if IsKnown("nitro") {
		t.Errorf("env-override should replace defaults, not merge")
	}
}

func TestNormalizeAndCheck(t *testing.T) {
	ResetForTest()
	os.Unsetenv("PULSE_KNOWN_PRODUCTS")

	if got := NormalizeAndCheck("Nitro"); got != "nitro" {
		t.Errorf("NormalizeAndCheck(\"Nitro\") = %q, want \"nitro\"", got)
	}
	// Unknown product não rejeita — apenas loga
	if got := NormalizeAndCheck("UnknownProduct"); got != "unknownproduct" {
		t.Errorf("NormalizeAndCheck should normalize even unknown products, got %q", got)
	}
}
