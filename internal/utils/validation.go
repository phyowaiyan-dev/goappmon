package utils

import (
	"errors"
	"fmt"
	"net/mail"
	"strings"
)

const (
	AdminNameMinLength = 3
	AdminNameMaxLength = 50
)

func ValidateAdminName(name string) error {
	trimmed := strings.TrimSpace(name)
	if len(trimmed) < AdminNameMinLength {
		return fmt.Errorf("admin name must be at least %d characters", AdminNameMinLength)
	}
	if len(trimmed) > AdminNameMaxLength {
		return fmt.Errorf("admin name must be at most %d characters", AdminNameMaxLength)
	}
	return nil
}

func ValidateEmail(email string) error {
	trimmed := strings.TrimSpace(email)
	if trimmed == "" {
		return errors.New("email is required")
	}
	if _, err := mail.ParseAddress(trimmed); err != nil {
		return errors.New("email format is invalid")
	}
	return nil
}
