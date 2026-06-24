package utils

import "testing"

func TestValidateAdminName(t *testing.T) {
	if err := ValidateAdminName("ab"); err == nil {
		t.Fatal("expected short name to fail")
	}
	if err := ValidateAdminName("abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz"); err == nil {
		t.Fatal("expected long name to fail")
	}
	if err := ValidateAdminName("Admin User"); err != nil {
		t.Fatalf("expected valid name, got %v", err)
	}
}

func TestValidateEmail(t *testing.T) {
	if err := ValidateEmail(""); err == nil {
		t.Fatal("expected empty email to fail")
	}
	if err := ValidateEmail("not-an-email"); err == nil {
		t.Fatal("expected invalid email to fail")
	}
	if err := ValidateEmail("admin@example.com"); err != nil {
		t.Fatalf("expected valid email, got %v", err)
	}
}
