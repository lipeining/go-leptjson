package leptjson

import (
	"testing"
)

func expectEQInt(t *testing.T, expect int, actual int) {
	if expect != actual {
		t.Errorf("parse events, expect: %v, actual: %v", expect, actual)
	}
}
func expectEQLeptType(t *testing.T, expect, actual LeptType) {
	if expect != actual {
		t.Errorf("parse types, expect: %v, actual: %v", expect, actual)
	}
}
func TestLeptParseNull(t *testing.T) {
	v := &LeptValue{LeptFALSE}
	expectEQInt(t, LeptParseOK, LeptParse(v, "null"))
	expectEQLeptType(t, LeptNULL, LeptGetType(v))
}
func TestLeptParseTrue(t *testing.T) {
	v := &LeptValue{LeptFALSE}
	expectEQInt(t, LeptParseOK, LeptParse(v, "true"))
	expectEQLeptType(t, LeptTRUE, LeptGetType(v))
}
func TestLeptParseFalse(t *testing.T) {
	v := &LeptValue{LeptFALSE}
	expectEQInt(t, LeptParseOK, LeptParse(v, "false"))
	expectEQLeptType(t, LeptFALSE, LeptGetType(v))
}

func TestParseExpectValue(t *testing.T) {
	v := &LeptValue{LeptFALSE}

	expectEQInt(t, LeptParseExpectValue, LeptParse(v, ""))
	expectEQLeptType(t, LeptNULL, LeptGetType(v))

	expectEQInt(t, LeptParseExpectValue, LeptParse(v, " "))
	expectEQLeptType(t, LeptNULL, LeptGetType(v))
}
func TestParseInvalidValue(t *testing.T) {
	v := &LeptValue{LeptFALSE}

	expectEQInt(t, LeptParseInvalidValue, LeptParse(v, "nul"))
	expectEQLeptType(t, LeptNULL, LeptGetType(v))

	expectEQInt(t, LeptParseInvalidValue, LeptParse(v, "?"))
	expectEQLeptType(t, LeptNULL, LeptGetType(v))
}
func TestParseRootNotSingular(t *testing.T) {
	v := &LeptValue{LeptFALSE}

	expectEQInt(t, LeptParseRootNotSingular, LeptParse(v, "null x"))
	expectEQLeptType(t, LeptNULL, LeptGetType(v))
}
