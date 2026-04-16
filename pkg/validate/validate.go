package validate

import (
	"github.com/go-playground/validator/v10"
)

// Validator is a shared validator instance.
var Validator *validator.Validate

func init() {
	Validator = validator.New(validator.WithRequiredStructEnabled())

	// Register custom validations here if needed.
	// Example:
	// _ = Validator.RegisterValidation("cameroon_phone", validateCameroonPhone)
}

// Struct validates a struct using the shared validator instance.
func Struct(s interface{}) error {
	return Validator.Struct(s)
}
