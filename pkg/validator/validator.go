package validator

import (
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
)

var (
	validate *validator.Validate
	trans    ut.Translator
)

func Init() {
	validate = validator.New()
	english := en.New()
	uni := ut.New(english, english)
	trans, _ = uni.GetTranslator("en")
	en_translations.RegisterDefaultTranslations(validate, trans)
}

func Validate(i interface{}) error {
	return validate.Struct(i)
}

func TranslateError(err error) []string {
	if err == nil {
		return nil
	}

	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return []string{err.Error()}
	}

	var errors []string
	for _, e := range validationErrors {
		errors = append(errors, e.Translate(trans))
	}
	return errors
}
				
func RegisterCustomValidation(tag string, fn validator.Func) error {
	return validate.RegisterValidation(tag, fn)
}
