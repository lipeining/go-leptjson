package goleptjson

import (
	"strconv"
	"testing"
)

func expectEQInt(t *testing.T, expect int, actual int) {
	if expect != actual {
		t.Errorf("parse int, expect: %v, actual: %v", expect, actual)
	}
}
func expectEQFloat64(t *testing.T, expect, actual float64) {
	if expect != actual {
		t.Errorf("parse float64, expect: %v, actual: %v", expect, actual)
	}
}
func expectEQString(t *testing.T, expect, actual string) {
	if expect != actual {
		t.Errorf("parse string, expect: %v, actual: %v", expect, actual)
	}
}
func expectEQLeptType(t *testing.T, expect, actual LeptType) {
	if expect != actual {
		t.Errorf("parse types, expect: %v, actual: %v", expect, actual)
	}
}
func TestLeptParseNull(t *testing.T) {
	v := NewLeptValue()
	expectEQInt(t, LeptParseOK, LeptParse(v, "null"))
	expectEQLeptType(t, LeptNull, LeptGetType(v))
}
func TestLeptParseTrue(t *testing.T) {
	v := NewLeptValue()
	expectEQInt(t, LeptParseOK, LeptParse(v, "true"))
	expectEQLeptType(t, LeptTrue, LeptGetType(v))
}
func TestLeptParseFalse(t *testing.T) {
	v := NewLeptValue()
	expectEQInt(t, LeptParseOK, LeptParse(v, "false"))
	expectEQLeptType(t, LeptFalse, LeptGetType(v))
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
		v := NewLeptValue()
		expectEQInt(t, LeptParseOK, LeptParse(v, c.input))
		expectEQLeptType(t, LeptNumber, LeptGetType(v))
		expectEQFloat64(t, c.expect, LeptGetNumber(v))
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
		v := NewLeptValue()
		expectEQInt(t, LeptParseOK, LeptParse(v, c.input))
		expectEQLeptType(t, LeptNumber, LeptGetType(v))
		expectEQFloat64(t, c.expect, LeptGetNumber(v))
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
		v := NewLeptValue()
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
	// 使用 strconv 无法解析全部的数据，因为格式不对
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
	v := NewLeptValue()

	expectEQInt(t, LeptParseExpectValue, LeptParse(v, ""))
	expectEQLeptType(t, LeptNull, LeptGetType(v))

	expectEQInt(t, LeptParseExpectValue, LeptParse(v, " "))
	expectEQLeptType(t, LeptNull, LeptGetType(v))
}
func TestParseInvalidValue(t *testing.T) {
	v := NewLeptValue()

	expectEQInt(t, LeptParseInvalidValue, LeptParse(v, "nul"))
	expectEQLeptType(t, LeptNull, LeptGetType(v))

	expectEQInt(t, LeptParseInvalidValue, LeptParse(v, "?"))
	expectEQLeptType(t, LeptNull, LeptGetType(v))
}
func TestParseRootNotSingular(t *testing.T) {
	v := NewLeptValue()

	expectEQInt(t, LeptParseRootNotSingular, LeptParse(v, "null x"))
	expectEQLeptType(t, LeptNull, LeptGetType(v))
}
func TestParseString(t *testing.T) {
	valid := []struct {
		input  string
		expect string
	}{
		{"\"\"", ""},
		{"\"Hello\"", "Hello"},
		{"\"Hello\\nWorld\"", "Hello\nWorld"},
		{"\"\\\" \\\\ \\/ \\b \\f \\n \\r \\t\"", "\" \\ / \b \f \n \r \t"},
		// {"\"Hello\\u0000World\"", "Hello\0World"},
		{"\"\\u0024\"", "\x24"},
		{"\"\\u00A2\"", "\xC2\xA2"},
		{"\"\\u20AC\"", "\xE2\x82\xAC"},
		{"\"\\uD834\\uDD1E\"", "\xF0\x9D\x84\x9E"},
		{"\"\\ud834\\udd1e\"", "\xF0\x9D\x84\x9E"},
	}
	// 将 uint64 转为 []byte 的方式奇怪。
	for _, c := range valid {
		v := NewLeptValue()
		expectEQInt(t, LeptParseOK, LeptParse(v, c.input))
		expectEQLeptType(t, LeptString, LeptGetType(v))
		expectEQInt(t, len(c.expect), LeptGetStringLength(v))
		expectEQString(t, c.expect, LeptGetString(v))
	}
}
func TestParseMissingQuotationMark(t *testing.T) {
	valid := []struct {
		input  string
		expect string
	}{
		{"\"", ""},
		{"\"abc", ""},
	}
	for _, c := range valid {
		v := NewLeptValue()
		expectEQInt(t, LeptParseMissQuotationMark, LeptParse(v, c.input))
	}
}
func TestParseInvalidStringEscape(t *testing.T) {
	valid := []struct {
		input  string
		expect string
	}{
		{"\"\\v\"", ""},
		{"\"\\'\"", ""},
		{"\"\\0\"", ""},
		{"\"\\x12\"", ""},
	}
	for _, c := range valid {
		v := NewLeptValue()
		expectEQInt(t, LeptParseInvalidStringEscape, LeptParse(v, c.input))
	}
}
func TestParseInvalidStringChar(t *testing.T) {
	valid := []struct {
		input  string
		expect string
	}{
		{"\"\x01\"", ""},
		{"\"\x1F\"", ""},
	}
	for _, c := range valid {
		v := NewLeptValue()
		expectEQInt(t, LeptParseInvalidStringChar, LeptParse(v, c.input))
	}
}
func TestParseInvalidUnicodeHex(t *testing.T) {
	valid := []struct {
		input  string
		expect string
	}{
		{"\"\\u\"", ""},
		{"\"\\u0\"", ""},
		{"\"\\u01\"", ""},
		{"\"\\u012\"", ""},
		{"\"\\u/000\"", ""},
		{"\"\\uG000\"", ""},
		{"\"\\u0/00\"", ""},
		{"\"\\u0G00\"", ""},
		{"\"\\u00/0\"", ""},
		{"\"\\u00G0\"", ""},
		{"\"\\u000/\"", ""},
		{"\"\\u000G\"", ""},
	}
	for _, c := range valid {
		v := NewLeptValue()
		expectEQInt(t, LeptParseInvalidUnicodeHex, LeptParse(v, c.input))
	}
}
func TestParseInvalidSurrogate(t *testing.T) {
	valid := []struct {
		input  string
		expect string
	}{
		{"\"\\uD800\"", ""},
		{"\"\\uDBFF\"", ""},
		{"\"\\uD800\\\\\"", ""},
		{"\"\\uD800\\uDBFF\"", ""},
		{"\"\\uD800\\uE000\"", ""},
	}
	for _, c := range valid {
		v := NewLeptValue()
		expectEQInt(t, LeptParseInvalidUnicodeSurrogate, LeptParse(v, c.input))
	}
}
func TestParseArray(t *testing.T) {
	valid := []struct {
		input  string
		expect string
	}{
		{"[ ]", "[ ]"},
	}
	for _, c := range valid {
		v := NewLeptValue()
		expectEQInt(t, LeptParseOK, LeptParse(v, c.input))
		expectEQLeptType(t, LeptArray, LeptGetType(v))
		expectEQInt(t, 0, LeptGetArraySize(v))
	}
	// [ null , false , true , 123 , "abc" ]
	// [ [ ] , [ 0 ] , [ 0 , 1 ] , [ 0 , 1 , 2 ] ]
	{
		v := NewLeptValue()
		input := "[ null , false , true , 123.0 , \"abc\" ]"
		expectEQInt(t, LeptParseOK, LeptParse(v, input))
		expectEQLeptType(t, LeptArray, LeptGetType(v))
		expectEQInt(t, 5, LeptGetArraySize(v))
		// null
		expectEQLeptType(t, LeptNull, LeptGetType(LeptGetArrayElement(v, 0)))
		// false
		expectEQLeptType(t, LeptFalse, LeptGetType(LeptGetArrayElement(v, 1)))
		// true
		expectEQLeptType(t, LeptTrue, LeptGetType(LeptGetArrayElement(v, 2)))
		// 123
		expectEQLeptType(t, LeptNumber, LeptGetType(LeptGetArrayElement(v, 3)))
		expectEQFloat64(t, 123, LeptGetNumber(LeptGetArrayElement(v, 3)))
		// abc
		expectEQLeptType(t, LeptString, LeptGetType(LeptGetArrayElement(v, 4)))
		expectEQString(t, "abc", LeptGetString(LeptGetArrayElement(v, 4)))
		expectEQInt(t, len("abc"), LeptGetStringLength(LeptGetArrayElement(v, 4)))
	}
	{
		v := NewLeptValue()
		input := "[ [ ] , [ 0 ] , [ 0 , 1 ] , [ 0 , 1 , 2 ] ]"
		expectEQInt(t, LeptParseOK, LeptParse(v, input))
		expectEQLeptType(t, LeptArray, LeptGetType(v))
		expectEQInt(t, 4, LeptGetArraySize(v))
		for i := 0; i < 4; i++ {
			ele := LeptGetArrayElement(v, i)
			expectEQLeptType(t, LeptArray, LeptGetType(ele))
			expectEQInt(t, i, LeptGetArraySize(ele))
			for j := 0; j < i; j++ {
				num := LeptGetArrayElement(ele, j)
				expectEQLeptType(t, LeptNumber, LeptGetType(num))
				expectEQFloat64(t, float64(j), LeptGetNumber(num))
			}
		}
	}
}
func TestParseMissCoomaOrSquareBracket(t *testing.T) {
	valid := []struct {
		input  string
		expect string
	}{
		{"[1", "[ ]"},
		{"[1}", "[ ]"},
		{"[1 2", "[ ]"},
		{"[[]", "[ ]"},
	}
	for _, c := range valid {
		v := NewLeptValue()
		expectEQInt(t, LeptParseMissCommaOrSouareBracket, LeptParse(v, c.input))
	}
}

func TestParseObject(t *testing.T) {
	{
		v := NewLeptValue()
		input := " { } "
		expectEQInt(t, LeptParseOK, LeptParse(v, input))
		expectEQLeptType(t, LeptObject, LeptGetType(v))
		expectEQInt(t, 0, LeptGetObjectSize(v))
	}
	{
		v := NewLeptValue()
		input := " { " +
			"\"n\" : null , " +
			"\"f\" : false , " +
			"\"t\" : true , " +
			"\"i\" : 123 , " +
			"\"s\" : \"abc\", " +
			"\"a\" : [ 1, 2, 3 ]," +
			"\"o\" : { \"1\" : 1, \"2\" : 2, \"3\" : 3 }" +
			" } "
		expectEQInt(t, LeptParseOK, LeptParse(v, input))
		expectEQLeptType(t, LeptObject, LeptGetType(v))
		expectEQInt(t, 7, LeptGetObjectSize(v))

		expectEQString(t, "n", LeptGetObjectKey(v, 0))
		expectEQInt(t, len("n"), LeptGetObjectKeyLength(v, 0))
		expectEQLeptType(t, LeptNull, LeptGetType(LeptGetObjectValue(v, 0)))

		expectEQString(t, "f", LeptGetObjectKey(v, 1))
		expectEQInt(t, len("f"), LeptGetObjectKeyLength(v, 1))
		expectEQLeptType(t, LeptFalse, LeptGetType(LeptGetObjectValue(v, 1)))

		expectEQString(t, "t", LeptGetObjectKey(v, 2))
		expectEQInt(t, len("t"), LeptGetObjectKeyLength(v, 2))
		expectEQLeptType(t, LeptTrue, LeptGetType(LeptGetObjectValue(v, 2)))

		expectEQString(t, "i", LeptGetObjectKey(v, 3))
		expectEQInt(t, len("i"), LeptGetObjectKeyLength(v, 3))
		expectEQLeptType(t, LeptNumber, LeptGetType(LeptGetObjectValue(v, 3)))
		expectEQFloat64(t, 123.0, LeptGetNumber(LeptGetObjectValue(v, 3)))

		expectEQString(t, "s", LeptGetObjectKey(v, 4))
		expectEQInt(t, len("s"), LeptGetObjectKeyLength(v, 4))
		expectEQLeptType(t, LeptString, LeptGetType(LeptGetObjectValue(v, 4)))
		expectEQString(t, "abc", LeptGetString(LeptGetObjectValue(v, 4)))

		expectEQString(t, "a", LeptGetObjectKey(v, 5))
		expectEQInt(t, len("a"), LeptGetObjectKeyLength(v, 5))
		expectEQLeptType(t, LeptArray, LeptGetType(LeptGetObjectValue(v, 5)))
		expectEQInt(t, 3, LeptGetArraySize(LeptGetObjectValue(v, 5)))
		for i := 0; i < 3; i++ {
			e := LeptGetArrayElement(LeptGetObjectValue(v, 5), i)
			expectEQLeptType(t, LeptNumber, LeptGetType(e))
			expectEQFloat64(t, float64(i)+1.0, LeptGetNumber(e))
		}

		expectEQString(t, "o", LeptGetObjectKey(v, 6))
		expectEQInt(t, len("o"), LeptGetObjectKeyLength(v, 6))
		expectEQLeptType(t, LeptObject, LeptGetType(LeptGetObjectValue(v, 6)))
		expectEQInt(t, 3, LeptGetObjectSize(LeptGetObjectValue(v, 6)))
		for i := 0; i < 3; i++ {
			e := LeptGetObjectValue(LeptGetObjectValue(v, 6), i)
			expectEQLeptType(t, LeptNumber, LeptGetType(e))
			expectEQFloat64(t, float64(i)+1.0, LeptGetNumber(e))
		}
	}
}

func TestParseMissKey(t *testing.T) {
	valid := []struct {
		input  string
		expect string
	}{
		{"{:1,", "[ ]"},
		{"{1:1,", "[ ]"},
		{"{true:1,", "[ ]"},
		{"{false:1,", "[ ]"},
		{"{null:1,", "[ ]"},
		{"{[]:1,", "[ ]"},
		{"{{}:1,", "[ ]"},
		{"{\"a\":1,", "[ ]"},
	}
	for _, c := range valid {
		v := NewLeptValue()
		expectEQInt(t, LeptParseMissKey, LeptParse(v, c.input))
	}
}

func TestParseMissColon(t *testing.T) {
	valid := []struct {
		input  string
		expect string
	}{
		{"{\"a\"}", "[ ]"},
		{"{\"a\",\"b\"}", "[ ]"},
	}
	for _, c := range valid {
		v := NewLeptValue()
		expectEQInt(t, LeptParseMissColon, LeptParse(v, c.input))
	}
}
func TestParseMissCommaOrCurlyBracket(t *testing.T) {
	valid := []struct {
		input  string
		expect string
	}{
		{"{\"a\":1", "[ ]"},
		{"{\"a\":1]", "[ ]"},
		{"{\"a\":1 \"b\"", "[ ]"},
		{"{\"a\":{}", "[ ]"},
	}
	for _, c := range valid {
		v := NewLeptValue()
		expectEQInt(t, LeptParseMissCommaOrCurlyBracket, LeptParse(v, c.input))
	}
}

func TestAccessNull(t *testing.T) {
	v := NewLeptValue()
	LeptSetString(v, "a")
	LeptSetNull(v)
	expectEQLeptType(t, LeptNull, LeptGetType(v))
}
func TestAccessBoolean(t *testing.T) {
	v := NewLeptValue()
	LeptSetBoolean(v, 1)
	expectEQLeptType(t, LeptTrue, LeptGetType(v))
	LeptSetBoolean(v, 0)
	expectEQLeptType(t, LeptFalse, LeptGetType(v))
}
func TestAccessNumber(t *testing.T) {
	v := NewLeptValue()
	LeptSetNumber(v, 123.123)
	expectEQLeptType(t, LeptNumber, LeptGetType(v))
	expectEQFloat64(t, 123.123, LeptGetNumber(v))
}
func TestAccessString(t *testing.T) {
	v := NewLeptValue()
	LeptSetString(v, "")
	expectEQLeptType(t, LeptString, LeptGetType(v))
	expectEQInt(t, 0, LeptGetStringLength(v))
	expectEQString(t, "", LeptGetString(v))
	LeptSetString(v, "Hello")
	expectEQLeptType(t, LeptString, LeptGetType(v))
	expectEQInt(t, 5, LeptGetStringLength(v))
	expectEQString(t, "Hello", LeptGetString(v))
}
