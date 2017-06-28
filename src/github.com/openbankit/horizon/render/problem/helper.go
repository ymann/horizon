package problem

import (
	"github.com/openbankit/horizon/log"
	"bytes"
	"encoding/json"
	"fmt"
	"golang.org/x/net/context"
	"reflect"
)

func ShouldBeProblem(a interface{}, options ...interface{}) string {
	var actual P
	switch a.(type) {
	case *bytes.Buffer:
		body := a.(*bytes.Buffer)
		expected := options[0].(P)

		Inflate(context.Background(), &expected)

		err := json.Unmarshal(body.Bytes(), &actual)

		if err != nil {
			return fmt.Sprintf("Could not unmarshal json into problem struct:\n%s\n", body.String())
		}
	case P:
		actual = a.(P)
	case *P:
		pointer := a.(*P)
		actual = *pointer
	case nil:
		return "nil was not expected"
	default:
		return fmt.Sprintf("Type %s was not expected", reflect.TypeOf(a).Name())
	}

	expected := options[0].(P)

	if expected.Type != "" && actual.Type != expected.Type {
		return fmt.Sprintf("Mismatched problem type: %s expected, got %s", expected.Type, actual.Type)
	}

	if expected.Status != 0 && actual.Status != expected.Status {
		return fmt.Sprintf("Mismatched problem status: %s expected, got %s", expected.Status, actual.Status)
	}

	// check extras for invalid field
	if len(options) > 1 {
		expectedName := options[1].(string)
		log.WithField("extras", actual.Extras).Debug("Got problem with extras")
		actualName, ok := actual.Extras["invalid_field"]
		if !ok {
			return fmt.Sprintf("Expected extras to have invalid_field")
		}
		if expectedName != actualName.(string) {
			return fmt.Sprintf("Mismatched problem invalid field: %s expected, got %s", expectedName, actualName)
		}

	}

	return ""
}
