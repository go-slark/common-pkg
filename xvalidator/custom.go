package xvalidator

import (
	"github.com/gin-gonic/gin/binding"
	"github.com/globalsign/mgo/bson"
	"gopkg.in/go-playground/validator.v9"
	"regexp"
)

func RegisterCustomValidator() {
	v := binding.Validator.Engine().(*validator.Validate)
	v.RegisterValidation("is-objectid", ValidateObjectId)
	v.RegisterValidation("ValidateMobile", ValidateMobile)
	v.RegisterValidation("ValidateIdCard", ValidateIdCard)
}

func ValidateObjectId(fl validator.FieldLevel) bool {
	if idStr, ok := fl.Field().Interface().(string); ok {
		if idStr == "" || bson.IsObjectIdHex(idStr) {
			return true
		}
	}
	return false
}

func ValidateMobile(fl validator.FieldLevel) bool {
	regularExp := "/(^(0[0-9]{2,3}\\-)?([2-9][0-9]{6,7})+(\\-[0-9]{1,4})?$)|(^((\\(\\d{3}\\))|(\\d{3}\\-))?(1[3578]\\d{9})$)|(^(400)-(\\d{3})-(\\d{4})(.)(\\d{1,4})$)|(^(400)-(\\d{3})-(\\d{4}$))/"
	regExp := regexp.MustCompile(regularExp)
	mobile, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}

	if regExp.MatchString(mobile) {
		return true
	}
	return false
}

func ValidateIdCard(fl validator.FieldLevel) bool {
	regularExp := "^(\\d{17})([0-9]|X|x)$"
	regExp := regexp.MustCompile(regularExp)
	mobile, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}

	if regExp.MatchString(mobile) {
		return true
	}
	return false
}
