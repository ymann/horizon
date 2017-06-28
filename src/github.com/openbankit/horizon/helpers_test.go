package horizon

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"

	"github.com/openbankit/horizon/render/problem"
	"github.com/openbankit/horizon/test"
)

func NewTestApp() *App {
	app, err := NewApp(test.NewTestConfig())

	if err != nil {
		log.Panic(err)
	}

	return app
}

func NewRequestHelper(app *App) test.RequestHelper {
	return test.NewRequestHelper(app.web.router)
}

func ShouldBePageOf(actual interface{}, options ...interface{}) string {
	body := actual.(*bytes.Buffer)
	expected := options[0].(int)

	var result map[string]interface{}
	err := json.Unmarshal(body.Bytes(), &result)

	if err != nil {
		return fmt.Sprintf("Could not unmarshal json:\n%s\n", body.String())
	}

	embedded, ok := result["_embedded"]

	if !ok {
		return "No _embedded key in response"
	}

	records, ok := embedded.(map[string]interface{})["records"]
	if !ok {
		return "No records key in _embedded"
	}

	recordsArray, ok := records.([]interface{})
	if !ok {
		return "Expected array, but got instance"
	}

	length := len(recordsArray)

	if length != expected {
		return fmt.Sprintf("Expected %d records in page, got %d", expected, length)
	}

	return ""
}

func ShouldBeProblem(a interface{}, options ...interface{}) string {
	return problem.ShouldBeProblem(a, options...)
}
