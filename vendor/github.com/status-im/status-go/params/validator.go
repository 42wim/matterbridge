package params

import (
	"gopkg.in/go-playground/validator.v9"
)

// NewValidator returns a new validator.Validate.
func NewValidator() *validator.Validate {
	return validator.New()
}
