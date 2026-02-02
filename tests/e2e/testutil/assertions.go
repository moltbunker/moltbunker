package testutil

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

// Assertions provides test assertion helpers
type Assertions struct {
	t *testing.T
}

// NewAssertions creates a new Assertions instance
func NewAssertions(t *testing.T) *Assertions {
	return &Assertions{t: t}
}

// Equal asserts that two values are equal
func (a *Assertions) Equal(expected, actual interface{}, msgAndArgs ...interface{}) {
	a.t.Helper()
	if !reflect.DeepEqual(expected, actual) {
		msg := formatMessage("Expected values to be equal", msgAndArgs...)
		a.t.Errorf("%s\n  Expected: %v\n  Actual:   %v", msg, expected, actual)
	}
}

// NotEqual asserts that two values are not equal
func (a *Assertions) NotEqual(expected, actual interface{}, msgAndArgs ...interface{}) {
	a.t.Helper()
	if reflect.DeepEqual(expected, actual) {
		msg := formatMessage("Expected values to not be equal", msgAndArgs...)
		a.t.Errorf("%s\n  Value: %v", msg, actual)
	}
}

// True asserts that a value is true
func (a *Assertions) True(value bool, msgAndArgs ...interface{}) {
	a.t.Helper()
	if !value {
		msg := formatMessage("Expected true, got false", msgAndArgs...)
		a.t.Error(msg)
	}
}

// False asserts that a value is false
func (a *Assertions) False(value bool, msgAndArgs ...interface{}) {
	a.t.Helper()
	if value {
		msg := formatMessage("Expected false, got true", msgAndArgs...)
		a.t.Error(msg)
	}
}

// Nil asserts that a value is nil
func (a *Assertions) Nil(value interface{}, msgAndArgs ...interface{}) {
	a.t.Helper()
	if !isNil(value) {
		msg := formatMessage("Expected nil", msgAndArgs...)
		a.t.Errorf("%s\n  Actual: %v", msg, value)
	}
}

// NotNil asserts that a value is not nil
func (a *Assertions) NotNil(value interface{}, msgAndArgs ...interface{}) {
	a.t.Helper()
	if isNil(value) {
		msg := formatMessage("Expected non-nil value", msgAndArgs...)
		a.t.Error(msg)
	}
}

// NoError asserts that an error is nil
func (a *Assertions) NoError(err error, msgAndArgs ...interface{}) {
	a.t.Helper()
	if err != nil {
		msg := formatMessage("Expected no error", msgAndArgs...)
		a.t.Errorf("%s\n  Error: %v", msg, err)
	}
}

// Error asserts that an error is not nil
func (a *Assertions) Error(err error, msgAndArgs ...interface{}) {
	a.t.Helper()
	if err == nil {
		msg := formatMessage("Expected an error", msgAndArgs...)
		a.t.Error(msg)
	}
}

// ErrorContains asserts that an error contains a specific message
func (a *Assertions) ErrorContains(err error, contains string, msgAndArgs ...interface{}) {
	a.t.Helper()
	if err == nil {
		msg := formatMessage("Expected an error containing: "+contains, msgAndArgs...)
		a.t.Error(msg)
		return
	}
	if !strings.Contains(err.Error(), contains) {
		msg := formatMessage("Expected error to contain: "+contains, msgAndArgs...)
		a.t.Errorf("%s\n  Actual error: %v", msg, err)
	}
}

// Contains asserts that a string contains a substring
func (a *Assertions) Contains(s, contains string, msgAndArgs ...interface{}) {
	a.t.Helper()
	if !strings.Contains(s, contains) {
		msg := formatMessage("Expected string to contain: "+contains, msgAndArgs...)
		a.t.Errorf("%s\n  Actual: %s", msg, s)
	}
}

// NotContains asserts that a string does not contain a substring
func (a *Assertions) NotContains(s, notContains string, msgAndArgs ...interface{}) {
	a.t.Helper()
	if strings.Contains(s, notContains) {
		msg := formatMessage("Expected string to not contain: "+notContains, msgAndArgs...)
		a.t.Errorf("%s\n  Actual: %s", msg, s)
	}
}

// Len asserts that a slice or map has a specific length
func (a *Assertions) Len(object interface{}, length int, msgAndArgs ...interface{}) {
	a.t.Helper()
	v := reflect.ValueOf(object)
	if v.Kind() != reflect.Slice && v.Kind() != reflect.Map && v.Kind() != reflect.Array && v.Kind() != reflect.String {
		a.t.Error("Len assertion requires slice, map, array, or string")
		return
	}
	if v.Len() != length {
		msg := formatMessage("Expected length mismatch", msgAndArgs...)
		a.t.Errorf("%s\n  Expected: %d\n  Actual: %d", msg, length, v.Len())
	}
}

// Empty asserts that a slice, map, or string is empty
func (a *Assertions) Empty(object interface{}, msgAndArgs ...interface{}) {
	a.t.Helper()
	v := reflect.ValueOf(object)
	if v.Kind() != reflect.Slice && v.Kind() != reflect.Map && v.Kind() != reflect.Array && v.Kind() != reflect.String {
		a.t.Error("Empty assertion requires slice, map, array, or string")
		return
	}
	if v.Len() != 0 {
		msg := formatMessage("Expected empty", msgAndArgs...)
		a.t.Errorf("%s\n  Length: %d", msg, v.Len())
	}
}

// NotEmpty asserts that a slice, map, or string is not empty
func (a *Assertions) NotEmpty(object interface{}, msgAndArgs ...interface{}) {
	a.t.Helper()
	v := reflect.ValueOf(object)
	if v.Kind() != reflect.Slice && v.Kind() != reflect.Map && v.Kind() != reflect.Array && v.Kind() != reflect.String {
		a.t.Error("NotEmpty assertion requires slice, map, array, or string")
		return
	}
	if v.Len() == 0 {
		msg := formatMessage("Expected non-empty", msgAndArgs...)
		a.t.Error(msg)
	}
}

// JSONEqual asserts that two JSON strings are equal
func (a *Assertions) JSONEqual(expected, actual string, msgAndArgs ...interface{}) {
	a.t.Helper()

	var expectedObj, actualObj interface{}
	if err := json.Unmarshal([]byte(expected), &expectedObj); err != nil {
		a.t.Errorf("Failed to parse expected JSON: %v", err)
		return
	}
	if err := json.Unmarshal([]byte(actual), &actualObj); err != nil {
		a.t.Errorf("Failed to parse actual JSON: %v", err)
		return
	}

	if !reflect.DeepEqual(expectedObj, actualObj) {
		msg := formatMessage("JSON values not equal", msgAndArgs...)
		a.t.Errorf("%s\n  Expected: %s\n  Actual: %s", msg, expected, actual)
	}
}

// Greater asserts that a > b
func (a *Assertions) Greater(val, threshold interface{}, msgAndArgs ...interface{}) {
	a.t.Helper()
	if !compareValues(val, threshold, ">") {
		msg := formatMessage("Expected greater than", msgAndArgs...)
		a.t.Errorf("%s\n  Value: %v\n  Threshold: %v", msg, val, threshold)
	}
}

// GreaterOrEqual asserts that a >= b
func (a *Assertions) GreaterOrEqual(val, threshold interface{}, msgAndArgs ...interface{}) {
	a.t.Helper()
	if !compareValues(val, threshold, ">=") {
		msg := formatMessage("Expected greater than or equal", msgAndArgs...)
		a.t.Errorf("%s\n  Value: %v\n  Threshold: %v", msg, val, threshold)
	}
}

// Less asserts that a < b
func (a *Assertions) Less(val, threshold interface{}, msgAndArgs ...interface{}) {
	a.t.Helper()
	if !compareValues(val, threshold, "<") {
		msg := formatMessage("Expected less than", msgAndArgs...)
		a.t.Errorf("%s\n  Value: %v\n  Threshold: %v", msg, val, threshold)
	}
}

// LessOrEqual asserts that a <= b
func (a *Assertions) LessOrEqual(val, threshold interface{}, msgAndArgs ...interface{}) {
	a.t.Helper()
	if !compareValues(val, threshold, "<=") {
		msg := formatMessage("Expected less than or equal", msgAndArgs...)
		a.t.Errorf("%s\n  Value: %v\n  Threshold: %v", msg, val, threshold)
	}
}

// formatMessage formats the assertion message
func formatMessage(defaultMsg string, msgAndArgs ...interface{}) string {
	if len(msgAndArgs) == 0 {
		return defaultMsg
	}
	if len(msgAndArgs) == 1 {
		if s, ok := msgAndArgs[0].(string); ok {
			return s
		}
	}
	if s, ok := msgAndArgs[0].(string); ok {
		return strings.TrimSpace(s + " - " + defaultMsg)
	}
	return defaultMsg
}

// isNil checks if a value is nil
func isNil(v interface{}) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return rv.IsNil()
	}
	return false
}

// compareValues compares two numeric values
func compareValues(a, b interface{}, op string) bool {
	aFloat, aOk := toFloat64(a)
	bFloat, bOk := toFloat64(b)
	if !aOk || !bOk {
		return false
	}

	switch op {
	case ">":
		return aFloat > bFloat
	case ">=":
		return aFloat >= bFloat
	case "<":
		return aFloat < bFloat
	case "<=":
		return aFloat <= bFloat
	}
	return false
}

// toFloat64 converts a numeric value to float64
func toFloat64(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case int:
		return float64(n), true
	case int8:
		return float64(n), true
	case int16:
		return float64(n), true
	case int32:
		return float64(n), true
	case int64:
		return float64(n), true
	case uint:
		return float64(n), true
	case uint8:
		return float64(n), true
	case uint16:
		return float64(n), true
	case uint32:
		return float64(n), true
	case uint64:
		return float64(n), true
	case float32:
		return float64(n), true
	case float64:
		return n, true
	}
	return 0, false
}
