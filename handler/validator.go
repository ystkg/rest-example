package handler

import (
	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/ja"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	ja_translations "github.com/go-playground/validator/v10/translations/ja"
)

type customValidator struct {
	validator  *validator.Validate
	translator ut.Translator
}

func newValidator(locale string) *customValidator {
	v := validator.New()

	var trans ut.Translator
	if locale == "ja" {
		trans, _ = ut.New(ja.New()).FindTranslator()
		ja_translations.RegisterDefaultTranslations(v, trans)
	} else {
		trans, _ = ut.New(en.New()).FindTranslator()
		en_translations.RegisterDefaultTranslations(v, trans)
	}

	return &customValidator{validator: v, translator: trans}
}

func (cv *customValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}
