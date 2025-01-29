package validator

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

var EmailRX = regexp.MustCompile(
	`^[a-zA-Z0-9.!#$%&'*+/=?^_` + "`" + `{|}~-]+@` +
		`([a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+` +
		`[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?$`,
)

// var PasswordRX = regexp.MustCompile("^(?=.*[a-z])(?=.*[A-Z])(?=.*\\d)(?=.*[!@#$%^&*(),.?\":{}|<>]).{8,20}$")

type Validator struct {
	NonFieldErrors []string
	FieldErrors    map[string]string
}

func (v *Validator) Valid() bool {
	return len(v.FieldErrors) == 0 && len(v.NonFieldErrors) == 0
}

func (v *Validator) AddNonFieldError(message string) {
	v.NonFieldErrors = append(v.NonFieldErrors, message)
}

func (v *Validator) AddFieldError(key, message string) {

	if v.FieldErrors == nil {
		v.FieldErrors = make(map[string]string)
	}

	if _, exists := v.FieldErrors[key]; !exists {
		v.FieldErrors[key] = message
	}
}

func (v *Validator) CheckField(ok bool, key, message string) {
	if !ok {
		v.AddFieldError(key, message)
	}
}

func NotBlank(value string) bool {
	return strings.TrimSpace(value) != ""
}

func MaxChars(value string, n int) bool {
	return utf8.RuneCountInString(value) <= n
}

func PermittedValue[T comparable](value T, permittedValues ...T) bool {
	for i := range permittedValues {
		if value == permittedValues[i] {
			return true
		}
	}
	return false
}

func MinChars(value string, n int) bool {
	return utf8.RuneCountInString(value) >= n
}

func Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}

func ComparePassword(currentPassword, newPassword string) bool {
	return currentPassword == newPassword
}
func ValidatePassword(password string) bool {
	if len(password) < 8 || len(password) > 20 {
		return false
	}
	hasLower := false
	hasUpper := false
	hasDigit := false
	hasSpecial := false

	for _, char := range password {
		if char >= 'a' && char <= 'z' {
			hasLower = true
		} else if char >= 'A' && char <= 'Z' {
			hasUpper = true
		} else if char >= '0' && char <= '9' {
			hasDigit = true
		} else if strings.ContainsRune("!@#$%^&*(),.?\":{}|<>", char) {
			hasSpecial = true
		}
	}
	return hasLower && hasUpper && hasDigit && hasSpecial
}
