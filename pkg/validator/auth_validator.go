package validator

import (
	"github.com/go-playground/validator/v10"
)

// RegisterAuthValidators mendaftarkan validator khusus auth ke validator instance.
// Dipanggil sekali saat bootstrap aplikasi (lihat cmd/server/main.go).
func RegisterAuthValidators(v *validator.Validate) {
	// Memastikan confirm_password sama dengan password (eqfield=Password).
	// Catatan: validator v10 sebenarnya sudah punya tag bawaan "eqfield",
	// registrasi ulang ini menjaga perilaku tetap eksplisit untuk DTO auth.
	v.RegisterValidation("eqfield", func(fl validator.FieldLevel) bool {
		field := fl.Field()
		otherField := fl.Parent().FieldByName(fl.Param())

		if !otherField.IsValid() {
			return false
		}

		return field.String() == otherField.String()
	})
}
