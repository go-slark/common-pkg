package xvalidator

import (
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/locales/zh"
	"github.com/go-playground/universal-translator"
	"gopkg.in/go-playground/validator.v9"
	zhTrans "gopkg.in/go-playground/validator.v9/translations/zh"
)

var validTran ut.Translator

func init() {
	validTran, _ = ut.New(zh.New()).GetTranslator("zh")
	_ = zhTrans.RegisterDefaultTranslations(binding.Validator.Engine().(*validator.Validate), validTran)
}

func translate(trans ut.Translator, fe validator.FieldError) string {
	msg, err := trans.T(fe.Tag(), fe.Field())
	if err != nil {
		return fe.(error).Error()
	}
	return msg
}

func registerTranslator(tag string, msg string) validator.RegisterTranslationsFunc {
	return func(trans ut.Translator) error {
		return trans.Add(tag, msg, false)
	}
}

func Error(err error) []string {
	ves, ok := err.(validator.ValidationErrors)
	if !ok {
		return nil
	}

	e := make([]string, 0, len(ves))
	for _, ve := range ves {
		t := ve.Translate(validTran)
		e = append(e, t)
	}
	return e
}
