package xvalidator

import (
	"reflect"
	"sync"

	"github.com/gin-gonic/gin/binding"
	"gopkg.in/go-playground/validator.v9"
)

func setValidatorToV9() {
	binding.Validator = new(defaultValidator)
}

type defaultValidator struct {
	once     sync.Once
	validate *validator.Validate
}

var _ binding.StructValidator = &defaultValidator{}

func (v *defaultValidator) ValidateStruct(obj interface{}) error {
	if kindOfData(obj) != reflect.Struct {
		return nil
	}
	v.lazyInit()
	return v.validate.Struct(obj)
}

func (v *defaultValidator) Engine() interface{} {
	v.lazyInit()
	return v.validate
}

func (v *defaultValidator) lazyInit() {
	v.once.Do(func() {
		v.validate = validator.New()
		v.validate.SetTagName("binding")
	})
}

func kindOfData(data interface{}) reflect.Kind {
	value := reflect.ValueOf(data)
	valueType := value.Kind()
	if valueType == reflect.Ptr {
		valueType = value.Elem().Kind()
	}
	return valueType
}
