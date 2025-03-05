package valid

import (
	"fmt"
	"net/mail"
	"regexp"
)

var (
	isValidUsername = regexp.MustCompile(`^[a-z0-9_]+$`).MatchString
	isValidFullname = regexp.MustCompile(`^[a-zA-Z\s]+$`).MatchString
)

func ValidateString(value string, minLength int, maxLength int) (err error) {
	if len(value) > maxLength || len(value) < minLength {
		return fmt.Errorf("must be between %d and %d characters", minLength, maxLength)
	}
	return nil
}
func ValidateUsername(value string) (err error) {
	ValidateString(value, 3, 20)

	if !isValidUsername(value) {
		return fmt.Errorf("must contain only lowercase letters, numbers, and underscores")
	}	
	return nil
}

func ValidatePassword(validate string) (err error) {
	return ValidateString(validate, 6, 20)
}

func ValidateEmail(value string) (err error) {
	ValidateString(value, 3, 100)

	if _, err := mail.ParseAddress(value); err!=nil{
		return fmt.Errorf("must be a valid email address")
	}
	return nil
}

func ValidateFullName(value string) (err error) {
	ValidateString(value, 3, 100)
	if !isValidFullname(value) {
		return fmt.Errorf("must contain only letters and spaces")
	}
	return nil
}