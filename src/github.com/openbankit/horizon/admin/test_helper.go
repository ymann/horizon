package admin

import (
	"fmt"
	"reflect"
)

func ShouldBeInvalidField(a interface{}, options ...interface{}) string {
	var actual InvalidFieldError
	switch a.(type) {
	case *InvalidFieldError:
		pointer := a.(*InvalidFieldError)
		actual = *pointer
	case InvalidFieldError:
		actual = a.(InvalidFieldError)
	default:
		return fmt.Sprintf("Mismatched error type: %s expected, got %s", "InvalidFieldError", reflect.TypeOf(a).String())
	}

	expected := options[0].(string)
	if actual.FieldName != expected {
		return fmt.Sprintf("Mismatched problem invalid field: %s expected, got %s", expected, actual.FieldName)
	}

	return ""
}
