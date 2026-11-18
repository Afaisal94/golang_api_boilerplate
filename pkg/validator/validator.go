package validator

import (
	"github.com/go-playground/validator/v10"
)

// FieldError adalah format error validasi per field.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// getErrorMsg memetakan tag validator ke pesan yang ramah.
func getErrorMsg(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "This field is required"
	case "min":
		return "Should be at least " + fe.Param() + " characters"
	case "max":
		return "Should be at most " + fe.Param() + " characters"
	case "email":
		return "Must be a valid email address"
	case "oneof":
		return "Must be one of: " + fe.Param()
	case "eqfield":
		return "Must match " + fe.Param() + " field"
	}
	return "Validation failed on '" + fe.Tag() + "'"
}

// Format mengubah validator.ValidationErrors menjadi []FieldError.
// Package ini murni — tidak mengimpor gin — sehingga bisa dipakai
// oleh lapisan mana pun tanpa mengikatnya ke framework HTTP.
// Binding request tetap tanggung jawab lapisan handler.
func Format(ve validator.ValidationErrors) []FieldError {
	out := make([]FieldError, len(ve))
	for i, fe := range ve {
		out[i] = FieldError{Field: fe.Field(), Message: getErrorMsg(fe)}
	}
	return out
}
