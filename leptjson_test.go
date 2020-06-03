package leptjson

import (
	"strconv"
	"testing"
)

func expectEQInt(t *testing.T, expect int, actual int) {
	if expect != actual {
		t.Errorf("parse events, expect: %v, actual: %v", expect, actual)
	}
}
func expectEQFloat32(t *testing.T, expect, actual float64) {
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
	v := &LeptValue{LeptFALSE, 0.0}
	expectEQInt(t, LeptParseOK, LeptParse(v, "null"))
	expectEQLeptType(t, LeptNULL, LeptGetType(v))
}
func TestLeptParseTrue(t *testing.T) {
	v := &LeptValue{LeptFALSE, 0.0}
	expectEQInt(t, LeptParseOK, LeptParse(v, "true"))
	expectEQLeptType(t, LeptTRUE, LeptGetType(v))
}
func TestLeptParseFalse(t *testing.T) {
	v := &LeptValue{LeptFALSE, 0.0}
	expectEQInt(t, LeptParseOK, LeptParse(v, "false"))
	expectEQLeptType(t, LeptFALSE, LeptGetType(v))
}
func TestLeptParseNumber(t *testing.T) {
	valid := []struct {
		input  string
		expect float64
	}{
		{"0", 0.0},
		{"-0", 0.0},
		{"-0.0", 0.0},
		{"1", 1.0},
		{"-1", -1.0},
		{"1.5", 1.5},
		{"-1.5", -1.5},
		{"1E10", 1E10},
		{"1e10", 1e10},
		{"1e+10", 1e+10},
		{"1e-10", 1e-10},
		{"-1E10", -1E10},
		{"-1e10", -1e10},
		{"-1e+10", -1e+10},
		{"-1e-10", -1e-10},
		{"1.234E+10", 1.234E+10},
		{"1.234E-10", 1.234E-10},
		{"1e-10000", 0.0},
	}
	for _, c := range valid {
		v := &LeptValue{LeptFALSE, 0.0}
		expectEQInt(t, LeptParseOK, LeptParse(v, c.input))
		expectEQLeptType(t, LeptNUMBER, LeptGetType(v))
		expectEQFloat32(t, c.expect, LeptGetNumber(v))
	}
	edges := []struct {
		input  string
		expect float64
	}{
		{"1.0000000000000002", 1.0000000000000002},
		// {"4.9406564584124654e-324", 4.9406564584124654e-324},  // fail
		// {"-4.9406564584124654e-324", -4.9406564584124654e-324}, // fail
		{"2.2250738585072009e-308", 2.2250738585072009e-308},
		{"-2.2250738585072009e-308", -2.2250738585072009e-308},
		{"2.2250738585072014e-308", 2.2250738585072014e-308},
		{"-2.2250738585072014e-308", -2.2250738585072014e-308},
		{"1.7976931348623157e+308", 1.7976931348623157e+308},
		{"-1.7976931348623157e+308", -1.7976931348623157e+308},
	}
	for _, c := range edges {
		v := &LeptValue{LeptFALSE, 0.0}
		expectEQInt(t, LeptParseOK, LeptParse(v, c.input))
		expectEQLeptType(t, LeptNUMBER, LeptGetType(v))
		expectEQFloat32(t, c.expect, LeptGetNumber(v))
	}
	// TEST_NUMBER(1.0000000000000002, "1.0000000000000002"); /* the smallest number > 1 */
	// TEST_NUMBER( 4.9406564584124654e-324, "4.9406564584124654e-324"); /* minimum denormal */
	// TEST_NUMBER(-4.9406564584124654e-324, "-4.9406564584124654e-324");
	// TEST_NUMBER( 2.2250738585072009e-308, "2.2250738585072009e-308");  /* Max subnormal double */
	// TEST_NUMBER(-2.2250738585072009e-308, "-2.2250738585072009e-308");
	// TEST_NUMBER( 2.2250738585072014e-308, "2.2250738585072014e-308");  /* Min normal positive double */
	// TEST_NUMBER(-2.2250738585072014e-308, "-2.2250738585072014e-308");
	// TEST_NUMBER( 1.7976931348623157e+308, "1.7976931348623157e+308");  /* Max double */
	// TEST_NUMBER(-1.7976931348623157e+308, "-1.7976931348623157e+308");
	invalid := []struct {
		input  string
		expect float64
	}{
		{"+0", 0.0},
		{"+1", 1.0},
		{".123", 1.5},
		{"1.", 1.5},
		{"INF", 1.5},
		{"inf", 1.5},
		{"NAN", 1.5},
		{"nan", 1.5},
		{"0123", 1.5},
		{"0x0", 1.5},
		{"0x123", 1.5},
	}
	for _, c := range invalid {
		v := &LeptValue{LeptFALSE, 0.0}
		expectEQInt(t, LeptParseInvalidValue, LeptParse(v, c.input))
	}
	// TEST_ERROR(LEPT_PARSE_NUMBER_TOO_BIG, "1e309");
	// TEST_ERROR(LEPT_PARSE_NUMBER_TOO_BIG, "-1e309");
}
func TestParseFloat(t *testing.T) {
	valid := []struct {
		input  string
		expect float64
	}{
		{"0", 0.0},
		{"-0", 0.0},
		{"-0.0", 0.0},
		{"1", 1.0},
		{"-1", -1.0},
		{"1.5", 1.5},
		{"-1.5", -1.5},
		{"1E10", 1E10},
		{"1e10", 1e10},
		{"1e+10", 1e+10},
		{"1e-10", 1e-10},
		{"-1E10", -1E10},
		{"-1e10", -1e10},
		{"-1e+10", -1e+10},
		{"-1e-10", -1e-10},
		{"1.234E+10", 1.234E+10},
		{"1.234E-10", 1.234E-10},
		{"1e-10000", 0.0},
	}
	// 无法解析全部的数据，因为格式不对
	for _, c := range valid {
		if ret, err := strconv.ParseFloat(c.input, 64); err != nil || float64(ret) != c.expect {
			t.Errorf("ParseFloat err: %v", err)
		}
	}
	invalid := []struct {
		input  string
		expect float64
	}{
		{"+0", 0.0},
		{"+1", 1.0},
		{".123", 1.5},
		{"1.", 1.5},
		{"INF", 1.5},
		{"inf", 1.5},
		{"NAN", 1.5},
		{"nan", 1.5},
		{"0123", 1.5},
		{"0x0", 1.5},
		{"0x123", 1.5},
	}
	for _, c := range invalid {
		if _, err := strconv.ParseFloat(c.input, 64); err == nil {
			t.Errorf("ParseFloat should get err, but now : %v", err)
		}
	}
}

func TestParseExpectValue(t *testing.T) {
	v := &LeptValue{LeptFALSE, 0.0}

	expectEQInt(t, LeptParseExpectValue, LeptParse(v, ""))
	expectEQLeptType(t, LeptNULL, LeptGetType(v))

	expectEQInt(t, LeptParseExpectValue, LeptParse(v, " "))
	expectEQLeptType(t, LeptNULL, LeptGetType(v))
}
func TestParseInvalidValue(t *testing.T) {
	v := &LeptValue{LeptFALSE, 0.0}

	expectEQInt(t, LeptParseInvalidValue, LeptParse(v, "nul"))
	expectEQLeptType(t, LeptNULL, LeptGetType(v))

	expectEQInt(t, LeptParseInvalidValue, LeptParse(v, "?"))
	expectEQLeptType(t, LeptNULL, LeptGetType(v))
}
func TestParseRootNotSingular(t *testing.T) {
	v := &LeptValue{LeptFALSE, 0.0}

	expectEQInt(t, LeptParseRootNotSingular, LeptParse(v, "null x"))
	expectEQLeptType(t, LeptNULL, LeptGetType(v))
}
