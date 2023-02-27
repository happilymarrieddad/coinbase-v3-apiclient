package utils

import (
	"fmt"
	"reflect"

	validatorV10 "github.com/go-playground/validator/v10"
)

var validator *validatorV10.Validate

func init() {
	validator = validatorV10.New()
}

// Validate - validates an object based on it's tags
func Validate(obj interface{}) (err error) {
	err = validator.Struct(obj)

	if err != nil {
		if errs, ok := err.(*validatorV10.InvalidValidationError); ok {
			name := errs.Type.Name()
			if errs.Type.Kind() == reflect.Ptr {
				name = errs.Type.Elem().Name()
			}

			return fmt.Errorf("%s is null", name)
		}
	}

	return
}
