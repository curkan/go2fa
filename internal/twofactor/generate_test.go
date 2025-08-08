package twofactor

import "testing"

func TestGenerateTOTP(t *testing.T) {
	code, exp := GenerateTOTP("MFRGGZDFMZTWQ2LK")
	if len(code) != 6 {
		t.Fatalf("code length must be 6, got %d (%s)", len(code), code)
	}
	if exp <= 0 {
		t.Fatalf("expiration must be positive")
	}
}
