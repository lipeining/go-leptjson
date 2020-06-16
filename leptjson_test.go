package goleptjson

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"
)

func expectEQBool(t *testing.T, expect, actual bool) {
	if expect != actual {
		t.Errorf("parse bool, expect: %v, actual: %v", expect, actual)
	}
}
func expectEQInt(t *testing.T, expect, actual int) {
	if expect != actual {
		t.Errorf("parse int, expect: %v, actual: %v", expect, actual)
	}
}
func expectEQLeptEvent(t *testing.T, expect, actual LeptEvent) {
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
	expectEQLeptEvent(t, LeptParseOK, LeptParse(v, "null"))
	expectEQLeptType(t, LeptNull, LeptGetType(v))
}
func TestLeptParseTrue(t *testing.T) {
	v := NewLeptValue()
	expectEQLeptEvent(t, LeptParseOK, LeptParse(v, "true"))
	expectEQLeptType(t, LeptTrue, LeptGetType(v))
}
func TestLeptParseFalse(t *testing.T) {
	v := NewLeptValue()
	expectEQLeptEvent(t, LeptParseOK, LeptParse(v, "false"))
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
		expectEQLeptEvent(t, LeptParseOK, LeptParse(v, c.input))
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
		expectEQLeptEvent(t, LeptParseOK, LeptParse(v, c.input))
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
		expectEQLeptEvent(t, LeptParseInvalidValue, LeptParse(v, c.input))
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

	expectEQLeptEvent(t, LeptParseExpectValue, LeptParse(v, ""))
	expectEQLeptType(t, LeptNull, LeptGetType(v))

	expectEQLeptEvent(t, LeptParseExpectValue, LeptParse(v, " "))
	expectEQLeptType(t, LeptNull, LeptGetType(v))
}
func TestParseInvalidValue(t *testing.T) {
	v := NewLeptValue()

	expectEQLeptEvent(t, LeptParseInvalidValue, LeptParse(v, "nul"))
	expectEQLeptType(t, LeptNull, LeptGetType(v))

	expectEQLeptEvent(t, LeptParseInvalidValue, LeptParse(v, "?"))
	expectEQLeptType(t, LeptNull, LeptGetType(v))
}
func TestParseRootNotSingular(t *testing.T) {
	v := NewLeptValue()

	expectEQLeptEvent(t, LeptParseRootNotSingular, LeptParse(v, "null x"))
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
		expectEQLeptEvent(t, LeptParseOK, LeptParse(v, c.input))
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
		expectEQLeptEvent(t, LeptParseMissQuotationMark, LeptParse(v, c.input))
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
		expectEQLeptEvent(t, LeptParseInvalidStringEscape, LeptParse(v, c.input))
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
		expectEQLeptEvent(t, LeptParseInvalidStringChar, LeptParse(v, c.input))
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
		expectEQLeptEvent(t, LeptParseInvalidUnicodeHex, LeptParse(v, c.input))
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
		expectEQLeptEvent(t, LeptParseInvalidUnicodeSurrogate, LeptParse(v, c.input))
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
		expectEQLeptEvent(t, LeptParseOK, LeptParse(v, c.input))
		expectEQLeptType(t, LeptArray, LeptGetType(v))
		expectEQInt(t, 0, LeptGetArraySize(v))
	}
	// [ null , false , true , 123 , "abc" ]
	// [ [ ] , [ 0 ] , [ 0 , 1 ] , [ 0 , 1 , 2 ] ]
	{
		v := NewLeptValue()
		input := "[ null , false , true , 123.0 , \"abc\" ]"
		expectEQLeptEvent(t, LeptParseOK, LeptParse(v, input))
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
		expectEQLeptEvent(t, LeptParseOK, LeptParse(v, input))
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
		expectEQLeptEvent(t, LeptParseMissCommaOrSouareBracket, LeptParse(v, c.input))
	}
}

func TestParseObject(t *testing.T) {
	{
		v := NewLeptValue()
		input := " { } "
		expectEQLeptEvent(t, LeptParseOK, LeptParse(v, input))
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
		expectEQLeptEvent(t, LeptParseOK, LeptParse(v, input))
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
		expectEQLeptEvent(t, LeptParseMissKey, LeptParse(v, c.input))
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
		expectEQLeptEvent(t, LeptParseMissColon, LeptParse(v, c.input))
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
		expectEQLeptEvent(t, LeptParseMissCommaOrCurlyBracket, LeptParse(v, c.input))
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

func TestAccessObject(t *testing.T) {
	o := NewLeptValue()
	for j := 0; j <= 5; j += 5 {
		LeptSetObject(o)
		expectEQInt(t, 0, LeptGetObjectSize(o))
		for i := 0; i < 10; i++ {
			key := 'a' + i
			v := NewLeptValue()
			LeptSetNumber(v, float64(i))
			LeptMove(LeptSetObjectValue(o, string(key)), v)
		}
		expectEQInt(t, 10, LeptGetObjectSize(o))
		for i := 0; i < 10; i++ {
			key := 'a' + i
			index := LeptFindObjectIndex(o, string(key))
			expectEQBool(t, true, index-LeptKeyNotExist != 0)
			pv := LeptGetObjectValue(o, index)
			expectEQFloat64(t, float64(i), LeptGetNumber(pv))
		}
	}
	{
		index := LeptFindObjectIndex(o, "j")
		expectEQBool(t, true, index-LeptKeyNotExist != 0)
		LeptRemoveObjectValue(o, index)
		index = LeptFindObjectIndex(o, "j")
		expectEQInt(t, index, LeptKeyNotExist)
		expectEQInt(t, 9, LeptGetObjectSize(o))
	}
	{
		index := LeptFindObjectIndex(o, "a")
		expectEQBool(t, true, index-LeptKeyNotExist != 0)
		LeptRemoveObjectValue(o, index)
		index = LeptFindObjectIndex(o, "a")
		expectEQInt(t, index, LeptKeyNotExist)
		expectEQInt(t, 8, LeptGetObjectSize(o))
	}
	{
		for i := 0; i < 8; i++ {
			key := 'a' + i + 1
			index := LeptFindObjectIndex(o, string(key))
			expectEQBool(t, true, index-LeptKeyNotExist != 0)
			pv := LeptGetObjectValue(o, index)
			expectEQFloat64(t, float64(i+1), LeptGetNumber(pv))
		}
	}
	{
		v := NewLeptValue()
		LeptSetString(v, "Hello")
		LeptMove(LeptSetObjectValue(o, "World"), v)
		pv := LeptFindObjectValue(o, "World")
		expectEQBool(t, true, pv != nil)
		expectEQString(t, "Hello", LeptGetString(pv))
	}
}
func TestLeptStringify(t *testing.T) {
	bases := []struct {
		input string
	}{
		{"null"},
		{"true"},
		{"false"},
	}
	for _, c := range bases {
		v := NewLeptValue()
		expectEQLeptEvent(t, LeptParseOK, LeptParse(v, c.input))
		actual := LeptStringify(v)
		expectEQString(t, c.input, actual)
	}

	numbers := []struct {
		input string
	}{
		{"0"},
		{"-0"},
		{"1"},
		{"-1"},
		{"1.5"},
		{"-1.5"},
		{"3.25"},
		{"1e+20"},
		{"1.234e+20"},
		{"1.234e-20"},

		{"1.0000000000000002"},
		// {"4.9406564584124654e-324"},
		// {"-4.9406564584124654e-324"},
		{"2.2250738585072009e-308"},
		{"-2.2250738585072009e-308"},
		{"2.2250738585072014e-308"},
		{"-2.2250738585072014e-308"},
		{"1.7976931348623157e+308"},
		{"-1.7976931348623157e+308"},
	}
	for _, c := range numbers {
		v := NewLeptValue()
		expectEQLeptEvent(t, LeptParseOK, LeptParse(v, c.input))
		actual := LeptStringify(v)
		expectEQString(t, c.input, actual)
		// 不完全正确的解析
	}

	strings := []struct {
		input string
	}{
		{"\"\""},
		{"\"Hello\""},
		{"\"Hello\\nWorld\""},
		{"\"\\\" \\\\ / \\b \\f \\n \\r \\t\""},
		{"\"Hello\\u0000World\""},
	}
	for _, c := range strings {
		v := NewLeptValue()
		expectEQLeptEvent(t, LeptParseOK, LeptParse(v, c.input))
		actual := LeptStringify(v)
		expectEQString(t, c.input, actual)
	}

	arrays := []struct {
		input string
	}{
		{"[]"},
		{"[null,false,true,123,\"abc\",[1,2,3]]"},
	}
	for _, c := range arrays {
		v := NewLeptValue()
		expectEQLeptEvent(t, LeptParseOK, LeptParse(v, c.input))
		actual := LeptStringify(v)
		expectEQString(t, c.input, actual)
	}

	objects := []struct {
		input string
	}{
		{"{}"},
		{"{\"n\":null,\"f\":false,\"t\":true,\"i\":123,\"s\":\"abc\",\"a\":[1,2,3],\"o\":{\"1\":1,\"2\":2,\"3\":3}}"},
	}
	for _, c := range objects {
		v := NewLeptValue()
		expectEQLeptEvent(t, LeptParseOK, LeptParse(v, c.input))
		actual := LeptStringify(v)
		expectEQString(t, c.input, actual)
	}
}

func TestLeptIsEqual(t *testing.T) {
	valid := []struct {
		inputLeft  string
		inputRight string
		expect     bool
	}{
		{"true", "true", true},
		{"true", "false", false},
		{"false", "false", true},
		{"null", "null", true},
		{"null", "0", false},
		{"123", "123", true},
		{"123", "456", false},
		{"\"abc\"", "\"abc\"", true},
		{"\"abc\"", "\"abcd\"", false},
		{"[]", "[]", true},
		{"[]", "null", false},
		{"[1,2,3]", "[1,2,3]", true},
		{"[1,2,3]", "[1,2,3,4]", false},
		{"[[]]", "[[]]", true},
		{"{}", "{}", true},
		{"{}", "null", false},
		{"{}", "[]", false},
		{"{\"a\":1,\"b\":2}", "{\"a\":1,\"b\":2}", true},
		{"{\"a\":1,\"b\":2}", "{\"b\":2,\"a\":1}", true},
		{"{\"a\":1,\"b\":2}", "{\"a\":1,\"b\":3}", false},
		{"{\"a\":1,\"b\":2}", "{\"a\":1,\"b\":2,\"c\":3}", false},
		{"{\"a\":{\"b\":{\"c\":{}}}}", "{\"a\":{\"b\":{\"c\":{}}}}", true},
		{"{\"a\":{\"b\":{\"c\":{}}}}", "{\"a\":{\"b\":{\"c\":[]}}}", false},
	}
	for _, c := range valid {
		vl, vr := NewLeptValue(), NewLeptValue()
		expectEQLeptEvent(t, LeptParseOK, LeptParse(vl, c.inputLeft))
		expectEQLeptEvent(t, LeptParseOK, LeptParse(vr, c.inputRight))
		expectEQBool(t, c.expect, LeptIsEqual(vl, vr))
	}
}
func TestLeptCopy(t *testing.T) {
	vl, vr := NewLeptValue(), NewLeptValue()
	LeptParse(vl, "{\"t\":true,\"f\":false,\"n\":null,\"d\":1.5,\"a\":[1,2,3]}")
	LeptCopy(vr, vl)
	expectEQBool(t, true, LeptIsEqual(vl, vr))
}
func TestLeptMove(t *testing.T) {
	vl, vr, vo := NewLeptValue(), NewLeptValue(), NewLeptValue()
	LeptParse(vl, "{\"t\":true,\"f\":false,\"n\":null,\"d\":1.5,\"a\":[1,2,3]}")
	LeptCopy(vr, vl)
	expectEQBool(t, true, LeptIsEqual(vl, vr))
	LeptMove(vo, vr)
	expectEQLeptType(t, LeptNull, LeptGetType(vr))
	expectEQBool(t, true, LeptIsEqual(vo, vl))
}
func TestLeptSwap(t *testing.T) {
	vl, vr := NewLeptValue(), NewLeptValue()
	LeptSetString(vl, "Hello")
	LeptSetString(vr, "World")
	LeptSwap(vl, vr)
	expectEQString(t, "World", LeptGetString(vl))
	expectEQString(t, "Hello", LeptGetString(vr))
}

func TestToMap(t *testing.T) {
	input := " { " +
		"\"n\" : null , " +
		"\"f\" : false , " +
		"\"t\" : true , " +
		"\"i\" : 123 , " +
		"\"s\" : \"abc\", " +
		"\"a\" : [ 1, 2, 3 ]," +
		"\"o\" : { \"1\" : 1, \"2\" : 2, \"3\" : 3 }" +
		" } "
	v := NewLeptValue()
	expectEQLeptEvent(t, LeptParseOK, LeptParse(v, input))
	// fmt.Println(ToInterface(v))
	// fmt.Println(ToMap(v))
	{
		i := ToInterface(v)
		if v, iok := i.(map[string]interface{}); !iok {
			t.Errorf("ToInterface expect map[string]interface{} to be ok")
		} else {
			size := len(v)
			if size != 7 {
				t.Errorf("map size to be 7")
			}
			if vi, ok := v["n"]; !ok {
				t.Errorf("map[n] should exist")
			} else if vi != nil {
				t.Errorf("map[n] should be nil")
			}
			if vi, ok := v["t"]; !ok {
				t.Errorf("map[t] should exist")
			} else {
				if viT, viok := vi.(bool); !viok || viT != true {
					t.Errorf("map[t] should be true")
				}
			}
			if vi, ok := v["f"]; !ok {
				t.Errorf("map[f] should exist")
			} else {
				if viT, viok := vi.(bool); !viok || viT != false {
					t.Errorf("map[f] should be false")
				}
			}
			if vi, ok := v["i"]; !ok {
				t.Errorf("map[i] should exist")
			} else {
				if viT, viok := vi.(float64); !viok || viT != 123 {
					t.Errorf("map[i] should be float64")
				}
			}
			if vi, ok := v["s"]; !ok {
				t.Errorf("map[s] should exist")
			} else {
				if viT, viok := vi.(string); !viok || viT != "abc" {
					t.Errorf("map[s] should be string")
				}
			}
			if vi, ok := v["a"]; !ok {
				t.Errorf("map[a] should exist")
			} else {
				if viT, viok := vi.([]interface{}); !viok || len(viT) != 3 {
					t.Errorf("map[a] should be array")
				}
			}
			if vi, ok := v["o"]; !ok {
				t.Errorf("map[o] should exist")
			} else {
				if viT, viok := vi.(map[string]interface{}); !viok || len(viT) != 3 {
					t.Errorf("map[o] should be object")
				}
			}
		}
	}
}
func TestToArray(t *testing.T) {
	input := "[null, true, false, 123, \"abc\", [ 1, 2, 3 ], { \"1\" : 1, \"2\" : 2, \"3\" : 3 }]"
	v := NewLeptValue()
	expectEQLeptEvent(t, LeptParseOK, LeptParse(v, input))
	// fmt.Println(ToInterface(v))
	// fmt.Println(ToArray(v))
	{
		i := ToInterface(v)
		if v, iok := i.([]interface{}); !iok {
			t.Errorf("ToInterface expect []interface{} to be ok")
		} else {
			size := len(v)
			if size != 7 {
				t.Errorf("array size to be 7")
			}
			if v[0] != nil {
				t.Errorf("v[0] should be nil")
			}
			if viT, viok := v[1].(bool); !viok || viT != true {
				t.Errorf("v[1] should be true")
			}
			if viT, viok := v[2].(bool); !viok || viT != false {
				t.Errorf("v[2] should be false")
			}
			if viT, viok := v[3].(float64); !viok || viT != 123 {
				t.Errorf("v[3] should be float64")
			}
			if viT, viok := v[4].(string); !viok || viT != "abc" {
				t.Errorf("v[4] should be string")
			}
			if viT, viok := v[5].([]interface{}); !viok || len(viT) != 3 {
				t.Errorf("v[5] should be array")
			}
			if viT, viok := v[6].(map[string]interface{}); !viok || len(viT) != 3 {
				t.Errorf("v[6] should be object")
			}
		}
	}
}
func TestToStruct(t *testing.T) {
	input := " { " +
		"\"N\" : null , " +
		"\"F\" : false , " +
		"\"T\" : true , " +
		"\"I\" : 123 , " +
		"\"S\" : \"abc\", " +
		"\"A\" : [ 1, 2, 3 ]," +
		"\"O\" : { \"1\" : 1, \"2\" : 2, \"3\" : 3 }" +
		" } "
	// input := " { " +
	// 	"\"n\" : null , " +
	// 	"\"f\" : false , " +
	// 	"\"t\" : true , " +
	// 	"\"i\" : 123 , " +
	// 	"\"s\" : \"abc\", " +
	// 	"\"a\" : [ 1, 2, 3 ]," +
	// 	"\"o\" : { \"1\" : 1, \"2\" : 2, \"3\" : 3 }" +
	// 	" } "
	v := NewLeptValue()
	expectEQLeptEvent(t, LeptParseOK, LeptParse(v, input))
	// fmt.Println(ToInterface(v))
	// fmt.Println(ToMap(v))
	type obj struct {
		N interface{}    `json:"n"`
		F bool           `json:"f"`
		T bool           `json:"t"`
		I int            `json:"i"`
		S string         `json:"s"`
		A []int          `json:"a"`
		O map[string]int `json:"o"`
	}
	{
		structure := obj{}
		err := ToStruct(v, &structure)
		if err != nil {
			t.Errorf("ToStruct expect no err: %v", err)
		} else {
			fmt.Println(structure)
			if vi := structure.N; vi != nil {
				t.Errorf("obj.N should be nil")
			}
			if vi := structure.F; vi != false {
				t.Errorf("obj.F should be false")
			}
			if vi := structure.T; vi != true {
				t.Errorf("obj.T should be true")
			}
			if vi := structure.I; vi != 123 {
				t.Errorf("obj.I should be 123")
			}
			if vi := structure.S; vi != "abc" {
				t.Errorf("obj.S should be \"abc\"")
			}
			if vi := structure.A; len(vi) != 3 {
				t.Errorf("obj.A should be slice and len = 3 ")
			} else {
				for j := 0; j < 3; j++ {
					if vi[j] != j+1 {
						t.Errorf("obj.A[%v] should be %v ", j, j+1)
					}
				}
			}
			if vi := structure.O; len(vi) != 3 {
				t.Errorf("obj.O should be map and len = 3 ")
			} else {
				for jindex, j := range []string{"1", "2", "3"} {
					if vi[j] != jindex+1 {
						t.Errorf("obj.O[%v] should be %v ", j, jindex+1)
					}
				}
			}
		}
	}
}
func TestToStructArray(t *testing.T) {
	input := "[null, true, false, 123, \"abc\", [ 1, 2, 3 ], { \"1\" : 1, \"2\" : 2, \"3\" : 3 }]"
	v := NewLeptValue()
	expectEQLeptEvent(t, LeptParseOK, LeptParse(v, input))
	// fmt.Println(ToInterface(v))
	// fmt.Println(ToArray(v))
	{
		var structure []interface{}
		if err := ToStruct(v, &structure); err != nil {
			t.Errorf("ToStruct expect no err: %v", err)
		} else {
			fmt.Println(structure)
			if vi := structure[0]; vi != nil {
				t.Errorf("array[0] should be nil")
			}
			if viT, viok := structure[1].(bool); !viok || viT != true {
				t.Errorf("structure[1] should be true")
			}
			if viT, viok := structure[2].(bool); !viok || viT != false {
				t.Errorf("structure[2] should be false")
			}
			if viT, viok := structure[3].(float64); !viok || viT != 123 {
				t.Errorf("structure[3] should be float64")
			}
			if viT, viok := structure[4].(string); !viok || viT != "abc" {
				t.Errorf("structure[4] should be string")
			}
			if viT, viok := structure[5].([]interface{}); !viok || len(viT) != 3 {
				t.Errorf("structure[5] should be array")
			} else {
				for j := 0; j < 3; j++ {
					if vitj, vijok := viT[j].(float64); !vijok || vitj != float64(j+1) {
						t.Errorf("array[0][%v] should be %v ", j, j+1)
					}
				}
			}
			// if viT, viok := structure[6].(map[string]interface{}); !viok || len(viT) != 3 {
			// 	t.Errorf("structure[6] should be object")
			// } else {
			// 	for jindex, j := range []string{"1", "2", "3"} {
			// 		if vitj, vijok := viT[j].(float64); !vijok || vitj != float64(jindex+1) {
			// 			t.Errorf("array[0][%v] should be %v ", j, jindex+1)
			// 		}
			// 	}
			// }
		}
	}
}

func TestSetValue(t *testing.T) {
	v := struct {
		A bool
	}{}
	// 检查 func(rv reflect.Value) 能否修改 rv 的值，这里是否是引用？
	// 结果是可以修改的，保证了递归的正确性
	rv := reflect.ValueOf(&v)
	rv = rv.Elem()
	setValue(rv.Field(0))
	fmt.Println(rv)
}
