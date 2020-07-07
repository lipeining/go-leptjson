package goleptjson

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"reflect"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"
)

var (
	// ErrReachEnd the input string is reach to end
	ErrReachEnd = errors.New("json string reach end")
	// ErrUnexpectChar expect function fail
	ErrUnexpectChar = errors.New("get an unexpect char")
)

// define some global parse events

// LeptEvent enums of parse event
type LeptEvent int

const (
	// LeptParseOK just ok
	LeptParseOK LeptEvent = iota
	// LeptParseExpectValue expect value
	LeptParseExpectValue
	// LeptParseInvalidValue invalid value
	LeptParseInvalidValue
	// LeptParseRootNotSingular root not singular
	LeptParseRootNotSingular

	// for number

	// LeptParseNumberTooBig number is to big
	LeptParseNumberTooBig

	// for string

	// LeptParseMissQuotationMark miss quotation mark
	LeptParseMissQuotationMark
	// LeptParseInvalidStringEscape invalid string escape
	LeptParseInvalidStringEscape
	// LeptParseInvalidStringChar invalid string char
	LeptParseInvalidStringChar
	// LeptParseInvalidUnicodeHex invalid unicode hex
	LeptParseInvalidUnicodeHex
	// LeptParseInvalidUnicodeSurrogate invalid unicode surrogate
	LeptParseInvalidUnicodeSurrogate

	// for array

	// LeptParseMissCommaOrSouareBracket miss comma or souare bracket
	LeptParseMissCommaOrSouareBracket

	// for object

	// LeptParseMissKey miss key
	LeptParseMissKey
	// LeptParseMissColon miss colon
	LeptParseMissColon
	// LeptParseMissCommaOrCurlyBracket miss cooma or curly bracket
	LeptParseMissCommaOrCurlyBracket
)

// LeptKeyNotExist object key not exist
const LeptKeyNotExist int = -1

// LeptType enums of json type
type LeptType int

const (
	// LeptNull parse to nil
	LeptNull LeptType = iota
	// LeptFalse parse to false
	LeptFalse
	// LeptTrue parse to true
	LeptTrue
	// LeptNumber parse to number like int float
	LeptNumber
	// LeptString parse to string
	LeptString
	// LeptArray parse to array
	LeptArray
	// LeptObject parse to map
	LeptObject
)

// LeptMember use to hold a pair of key/value
type LeptMember struct {
	key   string
	value *LeptValue
}

// LeptValue hold the value
type LeptValue struct {
	typ LeptType
	n   float64
	s   string
	a   []*LeptValue  // for array
	o   []*LeptMember // for object
}

// NewLeptValue return a init LeptValue
func NewLeptValue() *LeptValue {
	return &LeptValue{
		typ: LeptNull,
	}
}

// LeptContext hold the input string
type LeptContext struct {
	json string
}

// NewLeptContext return a init LeptContext
func NewLeptContext(json string) *LeptContext {
	return &LeptContext{
		json: json,
	}
}

func expect(c *LeptContext, ch byte) {
	if len(c.json) == 0 {
		panic(ErrReachEnd)
	}
	first := c.json[0]
	if first != ch {
		panic(ErrUnexpectChar)
	}
	c.json = c.json[1:]
}

// LeptParseWhitespace use to parse white space like '\t' '\n' '\r' ' '
func LeptParseWhitespace(c *LeptContext) {
	i := 0
	n := len(c.json)
	for i < n && (c.json[i] == ' ' || c.json[i] == '\t' || c.json[i] == '\n' || c.json[i] == '\r') {
		i++
	}
	c.json = c.json[i:]
}

// LeptParseNull use to parse "null"
func LeptParseNull(c *LeptContext, v *LeptValue) LeptEvent {
	expect(c, 'n')
	n := len(c.json)
	want := 4
	if n < want-1 {
		return LeptParseInvalidValue
	}
	if c.json[0] != 'u' || c.json[1] != 'l' || c.json[2] != 'l' {
		return LeptParseInvalidValue
	}
	c.json = c.json[want-1:]
	v.typ = LeptNull
	return LeptParseOK
}

// LeptParseTrue use to parse "true"
func LeptParseTrue(c *LeptContext, v *LeptValue) LeptEvent {
	expect(c, 't')
	n := len(c.json)
	want := 4
	if n < want-1 {
		return LeptParseInvalidValue
	}
	if c.json[0] != 'r' || c.json[1] != 'u' || c.json[2] != 'e' {
		return LeptParseInvalidValue
	}
	c.json = c.json[want-1:]
	v.typ = LeptTrue
	return LeptParseOK
}

// LeptParseFalse use to parse "false"
func LeptParseFalse(c *LeptContext, v *LeptValue) LeptEvent {
	expect(c, 'f')
	n := len(c.json)
	want := 5
	if n < want-1 {
		return LeptParseInvalidValue
	}
	if c.json[0] != 'a' || c.json[1] != 'l' || c.json[2] != 's' || c.json[3] != 'e' {
		return LeptParseInvalidValue
	}
	c.json = c.json[want-1:]
	v.typ = LeptFalse
	return LeptParseOK
}

// LeptParseLiteral merge null true false
func LeptParseLiteral(c *LeptContext, v *LeptValue, literal string, typ LeptType) LeptEvent {
	expect(c, literal[0])
	n := len(c.json)
	want := len(literal)
	if n < want-1 {
		return LeptParseInvalidValue
	}
	for i := 0; i < want-1; i++ {
		if c.json[i] != literal[i+1] {
			return LeptParseInvalidValue
		}
	}
	c.json = c.json[want-1:]
	v.typ = typ
	return LeptParseOK
}

// LeptParseNumber use to parse "Number"
func LeptParseNumber(c *LeptContext, v *LeptValue) LeptEvent {
	var end string
	var err error
	// v.n, end, err = strtod(c.json)
	v.n, end, err = strToFloat64(c.json)
	if err != nil {
		return LeptParseInvalidValue
	}
	c.json = end
	v.typ = LeptNumber
	return LeptParseOK
}

// strtod use to parse input string to a number
func strtod(input string) (float64, string, error) {
	// number = [ "-" ] int [ frac ] [ exp ]
	// int = "0" / digit1-9 *digit
	// frac = "." 1*digit
	// exp = ("e" / "E") ["-" / "+"] 1*digit

	// todo fix 浮点数溢出问题
	first := input[0]
	neg := false
	if first == '-' {
		neg = true
		input = input[1:]
	}
	var ret float64 = 0
	var integer int = 0
	var decimal int = 0
	var exp int = 0
	var err error
	n := len(input)
	var IllegalInput = errors.New("illegal input number string")
	if n == 0 {
		// no more charater
		return ret, "", IllegalInput
	}
	// take care of 0.0 0.12120
	if input[0] == '0' && n == 1 {
		// start with zero illegal like 0123
		return ret, "", nil
	}
	if input[0] == '0' && n > 1 {
		if input[1] == '.' {
			// pass have to check fix frac
		} else if input[1] == 'e' || input[1] == 'E' {
			// pass have to check fix exp
		} else if input[1] == 'x' || isDigit(input[1]) {
			// fix of 0x0 ox123 0123 0abc
			return ret, input, IllegalInput
		}
	}
	// fix 1abc 1, 1x
	if !isDigit(input[0]) {
		// 非法开头字符
		return ret, input, IllegalInput
	}
	input, integer, err = parseInteger(input)
	if err != nil {
		return ret, input, err
	}
	n = len(input)
	if n <= 0 {
		// end just integer
		ret = float64(integer)
		if neg {
			return -ret, input, nil
		}
		return ret, input, nil
	}
	// frac or exp
	ret = float64(integer)
	if input[0] == '.' {
		// should be frac
		input, decimal, err = parseFrac(input)
		if err != nil {
			return ret, input, err
		}
		var frac int = 1
		for j := n - 1; j > len(input); j-- {
			frac *= 10
		}
		// todo fix 浮点数溢出问题
		ret += float64(decimal) / float64(frac)
		if len(input) == 0 {
			if neg {
				return -ret, input, nil
			}
			return ret, input, nil
		}
		if !(input[0] == 'e' || input[0] == 'E') {
			// following is not exp  do not parse any more leave it to next parser
			return ret, input, nil
		}
		input, exp, err = parseExp(input)
		if err != nil {
			// illegal next char
			return ret, input, IllegalInput
		}
		ret *= float64(math.Pow10(exp))
		if neg {
			return -ret, input, nil
		}
		return ret, input, nil
	} else if input[0] == 'e' || input[0] == 'E' {
		// should be exp
		input, exp, err = parseExp(input)
		if err != nil {
			return ret, input, err
		}
		// do not parse any more leave it to next parser
		ret *= float64(math.Pow10(exp))
		if neg {
			return -ret, input, nil
		}
		return ret, input, nil
	} else {
		// illegal next like 123abc 123 , 123, 123] 123} 123space 等情况
		return ret, input, nil
	}
}

func strToFloat64(input string) (float64, string, error) {
	// number = [ "-" ] int [ frac ] [ exp ]
	// int = "0" / digit1-9 *digit
	// frac = "." 1*digit
	// exp = ("e" / "E") ["-" / "+"] 1*digit

	// fix 浮点数溢出问题
	origin := input
	end := 0
	if input[0] == '-' {
		input = input[1:]
		end++
	}
	var IllegalInput = errors.New("illegal input number string")
	if len(input) == 0 {
		// no more charater
		return 0, "", IllegalInput
	}
	// take care of 0.0 0.12120
	if input[0] == '0' && len(input) == 1 {
		// start with zero illegal like 0123
		return 0, "", nil
	}
	if input[0] == '0' && len(input) > 1 {
		if input[1] == '.' {
			// pass have to check fix frac
		} else if input[1] == 'e' || input[1] == 'E' {
			// pass have to check fix exp
		} else if input[1] == 'x' || isDigit(input[1]) {
			// fix of 0x0 ox123 0123 0abc
			return 0, input, IllegalInput
		}
	}
	// fix abc123 c1, x321
	if !isDigit(input[0]) {
		// 非法开头字符
		return 0, input, IllegalInput
	}
	i := 0
	for i < len(input) && isDigit(input[i]) {
		i++
	}
	input = input[i:]
	end += i
	if len(input) > 0 && input[0] == '.' {
		// should be frac
		if len(input) == 1 {
			return 0, input, IllegalInput
		}
		i = 1
		for i < len(input) && isDigit(input[i]) {
			i++
		}
		end += i
		input = input[i:]
		if len(input) > 0 && (input[0] == 'e' || input[0] == 'E') {
			if len(input) == 1 {
				return 0, input, IllegalInput
			}
			end++
			input = input[1:]
			if input[0] == '-' || input[0] == '+' {
				input = input[1:]
				end++
			}
			if len(input) == 0 {
				return 0, input, IllegalInput
			}
			i = 0
			for i < len(input) && isDigit(input[i]) {
				i++
			}
			end += i
			input = input[i:]
			ret, err := strconv.ParseFloat(origin[:end], 64)
			return ret, input, err
		}
		ret, err := strconv.ParseFloat(origin[:end], 64)
		return ret, input, err
	} else if len(input) > 0 && (input[0] == 'e' || input[0] == 'E') {
		if len(input) == 1 {
			return 0, input, IllegalInput
		}
		end++
		input = input[1:]
		if input[0] == '-' || input[0] == '+' {
			input = input[1:]
			end++
		}
		if len(input) == 0 {
			return 0, input, IllegalInput
		}
		i = 0
		for i < len(input) && isDigit(input[i]) {
			i++
		}
		end += i
		input = input[i:]
		ret, err := strconv.ParseFloat(origin[:end], 64)
		return ret, input, err
	} else {
		ret, err := strconv.ParseFloat(origin[:end], 64)
		return ret, input, err
	}
}

func parseExp(input string) (string, int, error) {
	if input[0] == 'e' || input[0] == 'E' {
		// should be exp
		if len(input) == 1 {
			// just e E illegal
			return "", 0, errors.New("input is not a exp")
		}
		expNeg := false
		if input[1] == '-' || input[1] == '+' {
			expNeg = input[1] == '-'
			input = input[2:]
		} else {
			input = input[1:]
		}
		// get exp
		input, exp, err := parseInteger(input)
		if err != nil {
			return "", 0, err
		}
		if expNeg {
			return input, -exp, err
		}
		return input, exp, err
	}
	return "", 0, errors.New("input is not a exp")
}
func parseFrac(input string) (string, int, error) {
	if len(input) == 1 {
		return "", 0, errors.New("input is not a frac")
	}
	if input[0] == '.' {
		// should be frac
		return parseInteger(input[1:])
	}
	return "", 0, errors.New("input is not a frac")
}
func parseInteger(input string) (string, int, error) {
	i := 0
	n := len(input)
	for i < n && isDigit(input[i]) {
		// get the integer
		i++
	}
	integer, err := strconv.Atoi(input[:i])
	if err != nil {
		return "", 0, err
	}
	return input[i:], integer, nil
}

func isDigit(char byte) bool {
	return char >= '0' && char <= '9'
}

func isDigit1to9(char byte) bool {
	return char >= '1' && char <= '9'
}

// LeptParseString use to parse string include \u
func LeptParseString(c *LeptContext, v *LeptValue) LeptEvent {
	s, ok := LeptParseStringRaw(c)
	if ok != LeptParseOK {
		return ok
	}
	LeptSetString(v, s)
	return ok
}

// LeptParseStringRaw use to parse string
// string = quotation-mark *char quotation-mark
// char = unescaped /
//    escape (
//        %x22 /          ; "    quotation mark  U+0022
//        %x5C /          ; \    reverse solidus U+005C
//        %x2F /          ; /    solidus         U+002F
//        %x62 /          ; b    backspace       U+0008
//        %x66 /          ; f    form feed       U+000C
//        %x6E /          ; n    line feed       U+000A
//        %x72 /          ; r    carriage return U+000D
//        %x74 /          ; t    tab             U+0009
//        %x75 4HEXDIG )  ; uXXXX                U+XXXX
// escape = %x5C          ; \
// quotation-mark = %x22  ; "
// unescaped = %x20-21 / %x23-5B / %x5D-10FFFF
func LeptParseStringRaw(c *LeptContext) (string, LeptEvent) {
	expect(c, '"')
	var stack bytes.Buffer
	defer stack.Truncate(0)
	for i, n := 0, len(c.json); i < n; i++ {
		ch := c.json[i]
		switch ch {
		case '"':
			c.json = c.json[i+1:]
			return stack.String(), LeptParseOK
		case '\\':
			// 遇到第一个转义符号，需要连续匹配两个 \
			if i+1 >= n {
				return "", LeptParseInvalidStringEscape
			}
			switch c.json[i+1] {
			case '"':
				stack.WriteString("\"")
			case '\\':
				stack.WriteString("\\")
			case 'b':
				stack.WriteString("\b")
			case 'f':
				stack.WriteString("\f")
			case 'n':
				stack.WriteString("\n")
			case 'r':
				stack.WriteString("\r")
			case 't':
				stack.WriteString("\t")
			case '/':
				stack.WriteString("/")
			// case 'u':
			// 	u, err := leptParseHex4(c.json[i+2:])
			// 	if err != nil {
			// 		return "", LeptParseInvalidUnicodeHex
			// 	}
			// 	if u < 0 || u > 0x10FFFF {
			// 		return "", LeptParseInvalidUnicodeHex
			// 	}
			// 	if u >= 0xD800 && u <= 0xDBFF { /* surrogate pair */
			// 		if i+6 >= n || c.json[i+6] != '\\' {
			// 			return "", LeptParseInvalidUnicodeSurrogate
			// 		}
			// 		if i+7 >= n || c.json[i+7] != 'u' {
			// 			return "", LeptParseInvalidUnicodeSurrogate
			// 		}
			// 		u2, err := leptParseHex4(c.json[i+8:])
			// 		if err != nil {
			// 			return "", LeptParseInvalidUnicodeHex
			// 		}
			// 		if u2 < 0xDC00 || u2 > 0xDFFF {
			// 			return "", LeptParseInvalidUnicodeSurrogate
			// 		}
			// 		u = (((u - 0xD800) << 10) | (u2 - 0xDC00)) + 0x10000
			// 		i += 6
			// 	}
			// 	检查代理对
			// 	stack.WriteString(leptEncodeUTF8(u))
			// 	if u <= 0x7F {
			// 		stack.Write(leptEncodeUTF8(u & 0xFF))
			// 	} else if u <= 0x7FF {
			// 		stack.Write(leptEncodeUTF8(0xC0 | ((u >> 6) & 0xFF)))
			// 		stack.Write(leptEncodeUTF8(0x80 | (u & 0x3F)))
			// 	} else if u <= 0xFFFF {
			// 		stack.Write(leptEncodeUTF8(0xE0 | ((u >> 12) & 0xFF)))
			// 		stack.Write(leptEncodeUTF8(0x80 | ((u >> 6) & 0x3F)))
			// 		stack.Write(leptEncodeUTF8(0x80 | (u & 0x3F)))
			// 	} else if u <= 0x10FFFF {
			// 		stack.Write(leptEncodeUTF8(0xF0 | ((u >> 18) & 0xFF)))
			// 		stack.Write(leptEncodeUTF8(0x80 | ((u >> 12) & 0x3F)))
			// 		stack.Write(leptEncodeUTF8(0x80 | ((u >> 6) & 0x3F)))
			// 		stack.Write(leptEncodeUTF8(0x80 | (u & 0x3F)))
			// 	} else {
			// 		panic("u is illegal")
			// 	}
			// 	// 将 uxxxx 跳过
			// 	// \\ 最后是有 i++ 这里只需要 4
			// 	i += 4
			case 'u':
				rr := getu4(c.json[i+2:])
				if rr < 0 {
					return "", LeptParseInvalidUnicodeHex
				}
				if utf16.IsSurrogate(rr) {
					if i+6 >= n || c.json[i+6] != '\\' {
						return "", LeptParseInvalidUnicodeSurrogate
					}
					if i+7 >= n || c.json[i+7] != 'u' {
						return "", LeptParseInvalidUnicodeSurrogate
					}
					rr1 := getu4(c.json[i+8:])
					if rr1 < 0xDC00 || rr1 > 0xDFFF {
						return "", LeptParseInvalidUnicodeSurrogate
					}
					if dec := utf16.DecodeRune(rr, rr1); dec != unicode.ReplacementChar {
						bits := make([]byte, 8)
						w := utf8.EncodeRune(bits, dec)
						stack.Write(bits[:w])
						i += 10
						// 这里的 break 是跳出 最近一层的 switch 所以需要加上下面的 i += 4
						break
					}
					// Invalid surrogate; fall back to replacement rune.
					rr = unicode.ReplacementChar
				}
				bits := make([]byte, 8)
				w := utf8.EncodeRune(bits, rr)
				stack.Write(bits[:w])
				i += 4
			default:
				return "", LeptParseInvalidStringEscape
			}
			// 这里的 i++ 针对普通的转码字符，至于 unicode 需要另外处理 uxxxx 个字符
			i++
		default:
			// 	unescaped = %x20-21 / %x23-5B / %x5D-10FFFF
			// 当中空缺的 %x22 是双引号，%x5C 是反斜线，都已经处理。所以不合法的字符是 %x00 至 %x1F。
			if ch < 0x20 {
				return "", LeptParseInvalidStringChar
			}
			stack.WriteByte(ch)
		}
	}
	// reach end of string becase the string has no \"
	return "", LeptParseMissQuotationMark
}

func leptEncodeUTF8(u uint64) []byte {
	bufSize := 8
	buf := make([]byte, bufSize)
	write := binary.PutUvarint(buf, u)
	// 这里奇怪 到底应该取 buf[:write] 还是 buf[:write-1]
	// todo fix \u0024 unicode encoding
	// 可能跟字节数有关，超过一定范围的数字就会有两个字节
	if write == 1 {
		return buf[:write]
	}
	return buf[:write-1]
}
func leptParseHex4(input string) (uint64, error) {
	n := len(input)
	if n < 4 {
		return 0, errors.New("illegal hex length of 4")
	}
	u, err := strconv.ParseUint(input[:4], 16, 64)
	if err != nil {
		return 0, errors.New("illegal hex string")
	}
	return u, nil
}
func getu4(input string) rune {
	n := len(input)
	if n < 4 {
		return -1
	}
	u, err := strconv.ParseUint(input[:4], 16, 64)
	if err != nil {
		return -1
	}
	return rune(u)
}

// LeptParseValue use to parse value switch to spec func
func LeptParseValue(c *LeptContext, v *LeptValue) LeptEvent {
	n := len(c.json)
	if n == 0 {
		return LeptParseExpectValue
	}
	switch c.json[0] {
	case 'n':
		return LeptParseNull(c, v)
	case 't':
		return LeptParseTrue(c, v)
	case 'f':
		return LeptParseFalse(c, v)
	case '"':
		return LeptParseString(c, v)
	case '[':
		return LeptParseArray(c, v)
	case '{':
		return LeptParseObject(c, v)
	default:
		return LeptParseNumber(c, v)
	}
}

// LeptParseArray use to parse array
func LeptParseArray(c *LeptContext, v *LeptValue) LeptEvent {
	// array = %x5B ws [ value *( ws %x2C ws value ) ] ws %x5D
	expect(c, '[')
	LeptParseWhitespace(c)
	n := len(c.json)
	if n == 0 {
		return LeptParseMissCommaOrSouareBracket
	}
	if c.json[0] == ']' {
		v.typ = LeptArray
		v.a = make([]*LeptValue, 0)
		c.json = c.json[1:]
		return LeptParseOK
	}
	for {
		// LeptParseWhitespace(c) // my
		vi := NewLeptValue()
		if ok := LeptParseValue(c, vi); ok != LeptParseOK {
			return ok
		}
		v.a = append(v.a, vi)
		// LeptParseWhitespace(c) //my
		// 教程中的解析 空格 时有道理的，需要在值之后解析 ws。具体参考对应的 regex 定义
		LeptParseWhitespace(c) // tutorial
		if len(c.json) == 0 {
			return LeptParseMissCommaOrSouareBracket
		}
		if c.json[0] == ',' {
			c.json = c.json[1:]
			LeptParseWhitespace(c) // tutorial
		} else if c.json[0] == ']' {
			c.json = c.json[1:]
			v.typ = LeptArray
			return LeptParseOK
		} else {
			return LeptParseMissCommaOrSouareBracket
		}
	}
}

// LeptParseObject use to parse object
func LeptParseObject(c *LeptContext, v *LeptValue) LeptEvent {
	// member = string ws %x3A ws value
	// object = %x7B ws [ member *( ws %x2C ws member ) ] ws %x7D
	expect(c, '{')
	LeptParseWhitespace(c)
	n := len(c.json)
	if n == 0 {
		return LeptParseMissCommaOrCurlyBracket
	}
	if c.json[0] == '}' {
		v.typ = LeptObject
		v.a = make([]*LeptValue, 0)
		v.o = make([]*LeptMember, 0)
		c.json = c.json[1:]
		return LeptParseOK
	}
	for {
		if len(c.json) == 0 || c.json[0] != '"' {
			return LeptParseMissKey
		}
		ki, ok := LeptParseStringRaw(c)
		if ok != LeptParseOK {
			return ok
		}
		// "":  23456789012E66, // fix 允许 key 为空字符串
		// if len(ki) == 0 {
		// 	return LeptParseMissKey
		// }
		LeptParseWhitespace(c)
		if c.json[0] != ':' {
			return LeptParseMissColon
		}
		c.json = c.json[1:]
		LeptParseWhitespace(c)
		vi := NewLeptValue()
		if ok := LeptParseValue(c, vi); ok != LeptParseOK {
			return ok
		}
		v.o = append(v.o, &LeptMember{key: ki, value: vi})
		// 教程中的解析 空格 时有道理的，需要在值之后解析 ws。具体参考对应的 regex 定义
		LeptParseWhitespace(c)
		if len(c.json) == 0 {
			return LeptParseMissCommaOrCurlyBracket
		}
		if c.json[0] == ',' {
			c.json = c.json[1:]
			LeptParseWhitespace(c)
		} else if c.json[0] == '}' {
			c.json = c.json[1:]
			v.typ = LeptObject
			return LeptParseOK
		} else {
			return LeptParseMissCommaOrCurlyBracket
		}
	}
}

// LeptParse use to parse value the enter
func LeptParse(v *LeptValue, json string) LeptEvent {
	if v == nil {
		panic("LeptParse v is nil")
	}
	c := NewLeptContext(json)
	v.typ = LeptNull
	LeptParseWhitespace(c)
	if ret := LeptParseValue(c, v); ret != LeptParseOK {
		return ret
	}
	LeptParseWhitespace(c)
	if len(c.json) != 0 {
		return LeptParseRootNotSingular
	}
	return LeptParseOK
}

// LeptGetType use to get the type of value
func LeptGetType(v *LeptValue) LeptType {
	if v == nil {
		panic("LeptGetType v is nil")
	}
	return v.typ
}

// LeptInit init v
func LeptInit(v *LeptValue) {
	v.typ = LeptNull
}

// LeptFree free the memory
func LeptFree(v *LeptValue) {
	// v = NewLeptValue()
	v.typ = LeptNull
	v.n = 0.0
	v.s = ""
	v.a = nil
	v.o = nil
}

// LeptSetNull use to set the type of null
func LeptSetNull(v *LeptValue) {
	if v == nil {
		panic("LeptGetNumber v is nil or typ is not LeptNumber")
	}
	v.typ = LeptNull
}

// LeptGetNumber use to get the type of value
func LeptGetNumber(v *LeptValue) float64 {
	if v == nil || v.typ != LeptNumber {
		panic("LeptGetNumber v is nil or typ is not LeptNumber")
	}
	return v.n
}

// LeptSetNumber use to set the type of value
func LeptSetNumber(v *LeptValue, n float64) {
	if v == nil {
		panic("LeptSetNumber v is nil ")
	}
	v.n = n
	v.typ = LeptNumber
}

// LeptGetBoolean use to get the type of value
func LeptGetBoolean(v *LeptValue) int {
	if v == nil || !(v.typ == LeptFalse || v.typ == LeptTrue) {
		panic("LeptGetBoolean v is nil or typ is not boolean")
	}
	if v.typ == LeptFalse {
		return 0
	}
	return 1
}

// LeptSetBoolean use to set the type of value
func LeptSetBoolean(v *LeptValue, n int) {
	if v == nil {
		panic("LeptSetBoolean v is nil ")
	}
	if n == 0 {
		v.typ = LeptFalse
	} else {
		v.typ = LeptTrue
	}
}

// LeptGetStringLength use to get the type of value
func LeptGetStringLength(v *LeptValue) int {
	if v == nil || v.typ != LeptString {
		panic("LeptGetStringLength v is nil or typ is not string")
	}
	return len(v.s)
}

// LeptGetString use to get the type of value
func LeptGetString(v *LeptValue) string {
	if v == nil || v.typ != LeptString {
		panic("LeptGetString v is nil or typ is not string")
	}
	return v.s
}

// LeptSetString use to get the type of value
func LeptSetString(v *LeptValue, s string) {
	if v == nil {
		panic("LeptSetString v is nil")
	}
	v.s = s
	v.typ = LeptString
}

// LeptGetArrayElement use to get the element of array[index]
func LeptGetArrayElement(v *LeptValue, index int) *LeptValue {
	if v == nil || v.typ != LeptArray {
		panic("LeptGetArrayElement v is nil or typ is not array")
	}
	if len(v.a) <= index {
		panic("LeptGetArrayElement v length <= input index")
	}
	return v.a[index]
}

// LeptGetArraySize use to get the size of array
func LeptGetArraySize(v *LeptValue) int {
	if v == nil || v.typ != LeptArray {
		panic("LeptGetArrayElement v is nil or typ is not array")
	}
	return len(v.a)
}

// LeptGetObjectSize use to get the size of object
func LeptGetObjectSize(v *LeptValue) int {
	if v == nil || v.typ != LeptObject {
		panic("LeptGetObjectSize v is nil or typ is not object")
	}
	return len(v.o)
}

// LeptGetObjectKey use to get the key of object
func LeptGetObjectKey(v *LeptValue, index int) string {
	if v == nil || v.typ != LeptObject {
		panic("LeptGetObjectKey v is nil or typ is not object")
	}
	if len(v.o) <= index {
		panic("LeptGetObjectKey v length <= input index")
	}
	member := v.o[index]
	return member.key
}

// LeptGetObjectKeyLength use to get the key length of object
func LeptGetObjectKeyLength(v *LeptValue, index int) int {
	if v == nil || v.typ != LeptObject {
		panic("LeptGetObjectKeyLength v is nil or typ is not object")
	}
	if len(v.o) <= index {
		panic("LeptGetObjectKeyLength v length <= input index")
	}
	member := v.o[index]
	return len(member.key)
}

// LeptGetObjectValue use to get the value of object
func LeptGetObjectValue(v *LeptValue, index int) *LeptValue {
	if v == nil || v.typ != LeptObject {
		panic("LeptGetObjectValue v is nil or typ is not object")
	}
	if len(v.o) <= index {
		panic("LeptGetObjectValue v length <= input index")
	}
	member := v.o[index]
	return member.value
}

// LeptStringify 得到紧凑的数据 string
func LeptStringify(v *LeptValue) string {
	return leptStringifyValue(v)
}

func leptStringifyValue(v *LeptValue) string {
	switch v.typ {
	case LeptNull:
		return "null"
	case LeptFalse:
		return "false"
	case LeptTrue:
		return "true"
	case LeptNumber:
		// return strconv.FormatFloat(v.n, 'g', -1, 64)
		return strconv.FormatFloat(v.n, 'g', 17, 64)
	case LeptString:
		return leptStringifyString(v.s)
	case LeptArray:
		return leptStringifyArray(v)
	case LeptObject:
		return leptStringifyObject(v)
	default:
		panic("leptStringifyValue invalid type")
	}
}

// leptStringifyString 考虑转义符号 unicode 字符集
func leptStringifyString(s string) string {
	var buf bytes.Buffer
	buf.WriteByte('"')
	hexDigits := []byte{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'A', 'B', 'C', 'D', 'E', 'F'}
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '"':
			buf.WriteByte('\\')
			buf.WriteByte('"')
		case '\\':
			buf.WriteByte('\\')
			buf.WriteByte('\\')
		case '\b':
			buf.WriteByte('\\')
			buf.WriteByte('b')
		case '\f':
			buf.WriteByte('\\')
			buf.WriteByte('f')
		case '\n':
			buf.WriteByte('\\')
			buf.WriteByte('n')
		case '\r':
			buf.WriteByte('\\')
			buf.WriteByte('r')
		case '\t':
			buf.WriteByte('\\')
			buf.WriteByte('t')
		default:
			if s[i] < 0x20 {
				buf.WriteByte('\\')
				buf.WriteByte('u')
				buf.WriteByte('0')
				buf.WriteByte('0')
				buf.WriteByte(hexDigits[s[i]>>4])
				buf.WriteByte(hexDigits[s[i]&15])
			} else {
				buf.WriteByte(s[i])
			}
		}
	}
	buf.WriteByte('"')
	return buf.String()
}

// leptStringifyArray 考虑转义符号 unicode 字符集
func leptStringifyArray(v *LeptValue) string {
	var buf bytes.Buffer
	buf.WriteByte('[')
	n := len(v.a)
	for i := 0; i < n; i++ {
		buf.WriteString(leptStringifyValue(v.a[i]))
		if i != n-1 {
			buf.WriteByte(',')
		}
	}
	buf.WriteByte(']')
	return buf.String()
}

// leptStringifyObject 考虑转义符号 unicode 字符集
func leptStringifyObject(v *LeptValue) string {
	var buf bytes.Buffer
	buf.WriteByte('{')
	n := len(v.o)
	for i := 0; i < n; i++ {
		key := v.o[i].key
		value := v.o[i].value
		buf.WriteString(leptStringifyString(key) + ":")
		buf.WriteString(leptStringifyValue(value))
		if i != n-1 {
			buf.WriteByte(',')
		}
	}
	buf.WriteByte('}')
	return buf.String()
}

// LeptCopy copy from src to dst
func LeptCopy(dst, src *LeptValue) bool {
	if dst == nil || src == nil {
		panic("src or dst is nil")
	}
	if dst == src {
		panic("src == dst")
	}
	switch src.typ {
	case LeptNull:
		LeptSetNull(dst)
	case LeptFalse:
		LeptSetBoolean(dst, 0)
	case LeptTrue:
		LeptSetBoolean(dst, 1)
	case LeptNumber:
		LeptSetNumber(dst, src.n)
	case LeptString:
		LeptSetString(dst, src.s)
	case LeptArray:
		for i := 0; i < len(src.a); i++ {
			ai := NewLeptValue()
			if ok := LeptCopy(ai, src.a[i]); !ok {
				return ok
			}
			dst.a = append(dst.a, ai)
		}
		dst.typ = LeptArray
	case LeptObject:
		for i := 0; i < len(src.o); i++ {
			oi := NewLeptValue()
			if ok := LeptCopy(oi, src.o[i].value); !ok {
				return ok
			}
			dst.o = append(dst.o, &LeptMember{key: src.o[i].key, value: oi})
		}
		dst.typ = LeptObject
	default:
		return false
	}
	return true
}

// LeptMove move from src to dst
func LeptMove(dst, src *LeptValue) bool {
	if dst == nil || src == nil {
		panic("src or dst is nil")
	}
	if dst == src {
		panic("src == dst")
	}
	LeptFree(dst)
	dst.typ = src.typ
	dst.n = src.n
	dst.s = src.s
	dst.a = src.a
	dst.o = src.o
	LeptFree(src)
	return true
}

// LeptSwap swap lhs rhs
func LeptSwap(lhs, rhs *LeptValue) bool {
	if lhs == nil || rhs == nil {
		panic("rhs or lhs is nil")
	}
	if lhs == rhs {
		panic("rhs == lhs")
	}
	lhs.typ, rhs.typ = rhs.typ, lhs.typ
	lhs.n, rhs.n = rhs.n, lhs.n
	lhs.s, rhs.s = rhs.s, lhs.s
	lhs.a, rhs.a = rhs.a, lhs.a
	lhs.o, rhs.o = rhs.o, lhs.o
	return true
}

// LeptIsEqual check lhs rhs is equal
func LeptIsEqual(lhs, rhs *LeptValue) bool {
	if lhs == nil || rhs == nil {
		panic("rhs or lhs is nil")
	}
	if lhs == rhs {
		return true
	}
	if lhs.typ != rhs.typ {
		return false
	}
	switch lhs.typ {
	case LeptNull:
		return true
	case LeptFalse:
		return true
	case LeptTrue:
		return true
	case LeptNumber:
		return lhs.n == rhs.n
	case LeptString:
		return lhs.s == rhs.s
	case LeptArray:
		if len(lhs.a) != len(rhs.a) {
			return false
		}
		for i := 0; i < len(lhs.a); i++ {
			if !LeptIsEqual(lhs.a[i], rhs.a[i]) {
				return false
			}
		}
		return true
	case LeptObject:
		if len(lhs.o) != len(rhs.o) {
			return false
		}
		for i := 0; i < len(lhs.o); i++ {
			key := lhs.o[i].key
			value := LeptFindObjectValue(rhs, key)
			if value == nil {
				return false
			}
			if !LeptIsEqual(lhs.o[i].value, value) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

// LeptFindObjectIndex find index
func LeptFindObjectIndex(v *LeptValue, key string) int {
	if v == nil || v.typ != LeptObject {
		panic("LeptFindObjectIndex v is nil or typ is not object")
	}
	for i := 0; i < len(v.o); i++ {
		if v.o[i].key == key {
			return i
		}
	}
	return LeptKeyNotExist
}

// LeptFindObjectValue find value
func LeptFindObjectValue(v *LeptValue, key string) *LeptValue {
	index := LeptFindObjectIndex(v, key)
	if index == LeptKeyNotExist {
		return nil
	}
	return v.o[index].value
}

// LeptSetObject set object value
func LeptSetObject(v *LeptValue) {
	if v == nil {
		panic("LeptSetObject v is nil")
	}
	LeptFree(v)
	v.o = make([]*LeptMember, 0)
	v.typ = LeptObject
}

// LeptSetObjectValue set object value
func LeptSetObjectValue(v *LeptValue, key string) *LeptValue {
	if v == nil || v.typ != LeptObject {
		panic("LeptSetObjectValue v is nil or typ is not object")
	}
	value := LeptFindObjectValue(v, key)
	if value != nil {
		return value
	}
	member := &LeptMember{key: key, value: NewLeptValue()}
	v.o = append(v.o, member)
	return member.value
}

// LeptRemoveObjectValue remove object value
func LeptRemoveObjectValue(v *LeptValue, index int) {
	if v == nil || v.typ != LeptObject {
		panic("LeptRemoveObjectValue v is nil or typ is not object")
	}
	size := len(v.o)
	if index >= size || index < 0 {
		panic("LeptRemoveObjectValue index >= size || index < 0")
	}
	// for i := index; i < size-1; i++ {
	// 	v.o[i] = v.o[i+1]
	// }
	// v.o = v.o[:size-1]
	next := make([]*LeptMember, size-1)
	copy(next, v.o[:index])
	copy(next, v.o[index+1:])
	v.o = next
}

// ToInterface transfer the LeptValue to golang interface{}
func ToInterface(v *LeptValue) interface{} {
	if v == nil {
		panic("ToInterface v is nil")
	}
	switch v.typ {
	case LeptNull:
		return nil
	case LeptFalse:
		return false
	case LeptTrue:
		return true
	case LeptNumber:
		return v.n
	case LeptString:
		return v.s
	case LeptArray:
		return ToArray(v)
	case LeptObject:
		return ToMap(v)
	default:
		panic("toInterface v typ error")
	}
}

// ToMap transafer the LeptValue to a golang map[string]interface
func ToMap(v *LeptValue) map[string]interface{} {
	if v == nil || v.typ != LeptObject {
		panic("ToMap v is nil or typ is not object")
	}
	size := len(v.o)
	m := make(map[string]interface{}, size)
	for i := 0; i < size; i++ {
		member := v.o[i]
		m[member.key] = ToInterface(member.value)
	}
	return m
}

// ToArray transafer the LeptValue to a golang []interface
func ToArray(v *LeptValue) []interface{} {
	if v == nil || v.typ != LeptArray {
		panic("ToArray v is nil or typ is not array")
	}
	size := len(v.a)
	arr := make([]interface{}, size)
	for i := 0; i < size; i++ {
		arr[i] = ToInterface(v.a[i])
	}
	return arr
}

// ToStruct transfer the LeptValue to a struct{} or []struct{}
func ToStruct(v *LeptValue, structure interface{}) error {
	rv := reflect.ValueOf(structure)
	if !rv.IsValid() {
		return fmt.Errorf("structure value is not valid")
	}
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("structure is not a ptr: %v", reflect.TypeOf(v))
	}
	return toValue(v, rv)
	// rv = rv.Elem()
	// 这里在对应的方法体内使用 indirect 处理 ptr
	// if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
	// 	return toSlice(v, rv)
	// } else if rv.Kind() == reflect.Struct {
	// 	return toStruct(v, rv)
	// } else if rv.Kind() == reflect.Map {
	// 	return toMap(v, rv)
	// }
	// return fmt.Errorf("structure value is not a ptr of slice or struct")
}

// Unmarshaler 导出的解析 json 的方法体
type Unmarshaler interface {
	// UnmarshalJSON get v and rv to set rv of
	UnmarshalJSON(v *LeptValue, rv reflect.Value) error
}

// 可以学习 encoding/json/decode.go
// 添加一个 indirect 方法，在里面进行不断地递归

// indirect walks down v allocating pointers as needed,
// until it gets to a non-pointer.
// if it encounters an Unmarshaler, indirect stops and returns that.
// if decodingNull is true, indirect stops at the last pointer so it can be set to nil.
func indirect(v reflect.Value, decodingNull bool) (Unmarshaler, reflect.Value) {
	// If v is a named type and is addressable,
	// start with its address, so that if the type has pointer methods,
	// we find them.
	if v.Kind() != reflect.Ptr && v.Type().Name() != "" && v.CanAddr() {
		v = v.Addr()
	}
	for {
		// Load value from interface, but only if the result will be
		// usefully addressable.
		if v.Kind() == reflect.Interface && !v.IsNil() {
			e := v.Elem()
			if e.Kind() == reflect.Ptr && !e.IsNil() && (!decodingNull || e.Elem().Kind() == reflect.Ptr) {
				v = e
				continue
			}
		}
		if v.Kind() != reflect.Ptr {
			break
		}
		if v.Elem().Kind() != reflect.Ptr && decodingNull && v.CanSet() {
			break
		}
		// fmt.Println(v, v.Type())
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		if v.Type().NumMethod() > 0 {
			if u, ok := v.Interface().(Unmarshaler); ok {
				return u, reflect.Value{}
			}
		}
		v = v.Elem()
	}
	return nil, v
}
func toValue(v *LeptValue, rv reflect.Value) error {
	// if rv.Kind() == reflect.Ptr {
	// 	rv = rv.Elem()
	// }
	if !rv.IsValid() {
		return fmt.Errorf("v is not valid")
	}
	// 针对自定义类型，需要判断是否有 UnmarshalJSON 方法
	decodingNull := v != nil && v.typ == LeptNull
	u, pv := indirect(rv, decodingNull)
	if u != nil {
		err := u.UnmarshalJSON(v, pv)
		return err
	}
	rv = pv
	if rv.Kind() == reflect.Array || rv.Kind() == reflect.Slice {
		return toSlice(v, rv)
	} else if rv.Kind() == reflect.Struct {
		return toStruct(v, rv)
	} else if rv.Kind() == reflect.Map {
		return toMap(v, rv)
	}
	// 这里开始，应该只有  bool, string, number
	// fmt.Println(rv.Type()) // goleptjson.LeptEvent
	// 可以传入自定义的 LeptEvent 对应的 Kind 还是包含在基本的 Kind 枚举中
	switch rv.Kind() {
	case reflect.Ptr:
		// 对应的 v 为 LeptNull 时， decodingNull = true
		fmt.Println("toValue got reflect.Ptr of v ", v)
	case reflect.Interface:
		// 可能对应的 rv 为 []interface{} interface{}
		// fmt.Println("toValue got reflect.Interface of v ", v, rv)
		if v == nil {
			rv.Set(reflect.Zero(rv.Type()))
		} else if rv.NumMethod() != 0 {
			return fmt.Errorf("umarshal v %v into type %v", v, rv.Type())
		} else {
			// 不能在 interface 上面进行各种 SetBool, SetFloat 操作
			switch v.typ {
			case LeptNull:
				// rv.Set(reflect.Zero(rv.Type()))
				// rv.Set(reflect.ValueOf(nil))
			case LeptFalse:
				rv.Set(reflect.ValueOf(false))
			case LeptTrue:
				rv.Set(reflect.ValueOf(true))
			case LeptNumber:
				rv.Set(reflect.ValueOf(v.n))
			case LeptString:
				rv.Set(reflect.ValueOf(v.s))
			case LeptArray:
				rvt := reflect.MakeSlice(reflect.SliceOf(rv.Type()), len(v.a), len(v.a))
				toSlice(v, rvt)
				rv.Set(rvt)
			case LeptObject:
				rvt := reflect.MakeMap(reflect.MapOf(reflect.TypeOf("abc"), rv.Type()))
				toMap(v, rvt)
				rv.Set(rvt)
			default:
				rv.Set(reflect.Zero(rv.Type()))
			}
		}
	case reflect.Bool:
		if v == nil {
			rv.SetBool(false)
		} else if v.typ == LeptFalse {
			rv.SetBool(false)
		} else if v.typ == LeptTrue {
			rv.SetBool(true)
		} else {
			return fmt.Errorf("v LeptValue is not a bool: %v", v.typ)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if v == nil {
			rv.SetInt(0)
		} else if v.typ == LeptNumber {
			rv.SetInt(int64(v.n))
		} else {
			return fmt.Errorf("v LeptValue is not a number: %v", v.typ)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if v == nil {
			rv.SetUint(0)
		} else if v.typ == LeptNumber {
			rv.SetUint(uint64(v.n))
		} else {
			return fmt.Errorf("v LeptValue is not a number: %v", v.typ)
		}
	case reflect.Float32, reflect.Float64:
		if v == nil {
			rv.SetFloat(0)
		} else if v.typ == LeptNumber {
			rv.SetFloat(float64(v.n))
		} else {
			return fmt.Errorf("v LeptValue is not a number: %v", v.typ)
		}
	case reflect.String:
		if v == nil {
			rv.SetString("")
		} else if v.typ == LeptString {
			rv.SetString(v.s)
		} else {
			return fmt.Errorf("v LeptValue is not a string: %v", v.typ)
		}
	default:
		// just ignore other Kind like chan, Func=
		if rv.IsValid() {
			rv.Set(reflect.Zero(rv.Type()))
		} else {
			switch v.typ {

			}
		}
	}
	// fmt.Println(rv)
	return nil
}
func parseTag(tag string) (string, string) {
	if idx := strings.Index(tag, ","); idx != -1 {
		return tag[:idx], tag[idx+1:]
	}
	return tag, ""
}
func toStruct(v *LeptValue, rv reflect.Value) error {
	if !rv.IsValid() {
		return fmt.Errorf("v is not valid")
	}
	decodingNull := v != nil && v.typ == LeptNull
	u, pv := indirect(rv, decodingNull)
	if u != nil {
		err := u.UnmarshalJSON(v, pv)
		return err
	}
	rv = pv
	size := rv.NumField()
	rt := rv.Type()
	// 这里没有考虑到 嵌套匿名字段 的处理
	for i := 0; i < size; i++ {
		fit := rt.Field(i)
		// fmt.Println(fit.Tag)
		tag := fit.Tag.Get("json")
		if tag == "-" {
			continue
		}
		name, opts := parseTag(tag)
		fiName := name
		// 只有 encode 的时候， omitempty 是起作用的
		if strings.Index(opts, "omitempty") != -1 {
			fmt.Println(tag, opts)
		}
		// fmt.Println()
		// fmt.Println(fit.Tag.Get("omitempty"))
		// fiName := fit.Name
		if v == nil {
			// if err := toValue(v, rv.Field(i)); err != nil {
			// 	return err
			// }
		} else if v.typ != LeptObject {
			return fmt.Errorf("v LeptValue is not a object: %v", v.typ)
		} else {
			liv := LeptFindObjectValue(v, fiName)
			if err := toValue(liv, rv.Field(i)); err != nil {
				return err
			}
		}
	}
	return nil
}
func toMap(v *LeptValue, rv reflect.Value) error {
	if !rv.IsValid() {
		return fmt.Errorf("v is not valid")
	}
	decodingNull := v != nil && v.typ == LeptNull
	u, pv := indirect(rv, decodingNull)
	if u != nil {
		err := u.UnmarshalJSON(v, pv)
		return err
	}
	rv = pv
	vsize := 0
	if v == nil {
	} else if v.typ != LeptObject {
		return fmt.Errorf("v LeptValue is not a object: %v", v.typ)
	} else {
		vsize = LeptGetObjectSize(v)
	}
	// fix panic: assignment to entry in nil map [recovered]
	if rv.IsNil() {
		rv.Set(reflect.MakeMap(rv.Type()))
	}
	// encoding/json/decode.go 如何处理 map[key]value 的 key
	// // Write value back to map;
	// // if using struct, subv points into struct already.
	// if v.Kind() == reflect.Map {
	// 	kv := reflect.ValueOf(key).Convert(v.Type().Key())
	// 	v.SetMapIndex(kv, subv)
	// }
	for i := 0; i < vsize; i++ {
		lik := LeptGetObjectKey(v, i)
		liv := LeptGetObjectValue(v, i)
		// 对于 key, value ，需要保证类型对应和递归处理
		// rikt := reflect.ValueOf(lik)
		// rivt := rv.Type().Elem() // get the value type
		// rivv := reflect.Zero(rivt)
		// if err := toValue(liv, rivv); err != nil {
		// 	return err
		// }
		// rv.SetMapIndex(rikt, rivv)
		// rikt := rv.Type().Key()
		// rikv := reflect.New(rikt).Elem()
		// rikv.Set(reflect.ValueOf(lik))
		rikv := reflect.ValueOf(lik)
		// 这里的 key 应该是 []byte
		// value of type []uint8 cannot be converted to type int
		// rikv := reflect.ValueOf([]byte(lik)).Convert(rv.Type().Key())
		// rikv := reflect.ValueOf(bytes.NewBufferString(lik).Bytes()).Convert(rv.Type().Key())
		rivt := rv.Type().Elem() // get the value type
		// 很可能对应的 elemType 并没有值，只是 Zero nil
		// rivp := reflect.New(rivt)
		// 可以考虑生成一个 Pointer，但是这部分的逻辑应该放在
		// toValue 中解决， toValue 中应该处理 Kind() ptr
		// elemType := v.Type().Elem()
		// 		if !mapElem.IsValid() {
		// 			mapElem = reflect.New(elemType).Elem()
		// 		} else {
		// 			mapElem.Set(reflect.Zero(elemType))
		// 		}
		var rivv reflect.Value
		// rivv.Set(reflect.Zero(rivt))
		rivv = reflect.New(rivt).Elem()
		if err := toValue(liv, rivv); err != nil {
			return err
		}
		rv.SetMapIndex(rikv, rivv)
	}
	return nil
}
func toSlice(v *LeptValue, rv reflect.Value) error {
	if !rv.IsValid() {
		return fmt.Errorf("v is not valid")
	}
	decodingNull := v != nil && v.typ == LeptNull
	u, pv := indirect(rv, decodingNull)
	if u != nil {
		err := u.UnmarshalJSON(v, pv)
		return err
	}
	rv = pv
	size := rv.Len()
	vsize := 0
	if v == nil {
		size = 0
	} else if v.typ != LeptArray {
		return fmt.Errorf("v LeptValue is not a array: %v", v.typ)
	} else {
		vsize = LeptGetArraySize(v)
	}
	// 没必要将 size 固定为输入，可能存在空的情况
	// 如果是 slice，可以设置对应的 length
	// fmt.Println(rv.Cap(), rv.Len(), rv.CanSet()) 0 0 true
	if vsize > rv.Cap() {
		newcap := vsize
		if newcap < 4 {
			newcap = 4
		}
		newv := reflect.MakeSlice(rv.Type(), rv.Len(), newcap)
		reflect.Copy(newv, rv)
		rv.Set(newv)
	}
	if size < vsize {
		rv.SetLen(vsize)
		size = vsize
	}
	// 这里或许应该以 v 为标准，因为传入的 rv 可能为 cap=0 的 slice
	for i := 0; i < size; i++ {
		if v == nil {
			// if err := toValue(v, rv.Index(i)); err != nil {
			// 	return err
			// }
		} else if v.typ != LeptArray {
			return fmt.Errorf("v LeptValue is not a array: %v", v.typ)
		} else {
			var liv *LeptValue
			if i < vsize {
				liv = LeptGetArrayElement(v, i)
			}
			if err := toValue(liv, rv.Index(i)); err != nil {
				return err
			}
		}
	}
	return nil
}
func setValue(rv reflect.Value) {
	rv.SetBool(true)
}

// Unmarshal parse input data into structure
func Unmarshal(data []byte, structure interface{}) error {
	v := NewLeptValue()
	event := LeptParse(v, string(data))
	if event != LeptParseOK {
		return fmt.Errorf("Unmarshal parse error: %v", event)
	}
	return ToStruct(v, structure)
}

// Marshaler is the interface implemented by objects that
// can marshal themselves into valid JSON.
type Marshaler interface {
	MarshalJSON() ([]byte, error)
}

var (
	marshalerType = reflect.TypeOf(new(Marshaler)).Elem()
)

type encodeState struct {
	bytes.Buffer
}

// Marshal stringify the input structure
func Marshal(structure interface{}) ([]byte, error) {
	e := &encodeState{}
	err := e.marshal(structure)
	if err != nil {
		return nil, err
	}
	return e.Bytes(), nil
}

func (e *encodeState) marshal(structure interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(runtime.Error); ok {
				panic(r)
			}
			if s, ok := r.(string); ok {
				panic(s)
			}
			err = r.(error)
		}
	}()
	e.reflectValue(reflect.ValueOf(structure), true)
	return nil
}
func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return false
}
func (e *encodeState) reflectValue(v reflect.Value, allowAddr bool) {
	if !v.IsValid() {
		e.WriteString("null")
		return
	}
	t := v.Type()
	if t.Implements(marshalerType) {
		marshalerEncoder(e, v, false)
		return
	}
	if t.Kind() != reflect.Ptr && allowAddr {
		if reflect.PtrTo(t).Implements(marshalerType) {
			if v.CanAddr() {
				addrMarshalerEncoder(e, v, false)
			} else {
				e.reflectValue(v, false)
			}
		}
		return
	}

	switch t.Kind() {
	case reflect.Bool:
		if v.Bool() {
			e.WriteString("true")
		} else {
			e.WriteString("false")
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		b := strconv.AppendInt([]byte(""), v.Int(), 10)
		e.Write(b)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		b := strconv.AppendUint([]byte(""), v.Uint(), 10)
		e.Write(b)
	case reflect.Float32:
		b := strconv.AppendFloat([]byte(""), v.Float(), 'g', -1, 32)
		e.Write(b)
	case reflect.Float64:
		b := strconv.AppendFloat([]byte(""), v.Float(), 'g', -1, 64)
		e.Write(b)
	case reflect.String:
		e.WriteString(leptStringifyString(v.String()))
	case reflect.Interface:
		if v.IsNil() {
			e.WriteString("null")
			return
		}
		e.reflectValue(v.Elem(), false)
	case reflect.Struct:
		e.WriteByte('{')
		first := true
		size := v.NumField()
		rt := t
		for i := 0; i < size; i++ {
			fit := rt.Field(i)
			// fmt.Println(fit.Tag)
			tag := fit.Tag.Get("json")
			if tag == "-" {
				continue
			}
			name, opts := parseTag(tag)
			// 只有 encode 的时候， omitempty 是起作用的
			fi := v.Field(i)
			if !fi.IsValid() || strings.Index(opts, "omitempty") != -1 && isEmptyValue(fi) {
				continue
			}
			if first {
				first = false
			} else {
				e.WriteByte(',')
			}
			e.WriteString(leptStringifyString(name))
			e.WriteByte(':')
			e.reflectValue(fi, true)
		}
		e.WriteByte('}')
	case reflect.Map:
		if v.IsNil() {
			e.WriteString("null")
			return
		}
		e.WriteByte('{')
		sv := v.MapKeys()
		sort.Slice(sv, func(i, j int) bool {
			return sv[i].String() < sv[j].String()
		})
		for i, k := range sv {
			if i > 0 {
				e.WriteByte(',')
			}
			e.WriteString(leptStringifyString(k.String()))
			e.WriteByte(':')
			// me.elemEnc(e, v.MapIndex(k), false)
			e.reflectValue(v.MapIndex(k), false)
		}
		e.WriteByte('}')
	case reflect.Slice:
		if v.IsNil() {
			e.WriteString("null")
			return
		}
		fallthrough
	case reflect.Array:
		e.WriteByte('[')
		n := v.Len()
		for i := 0; i < n; i++ {
			if i > 0 {
				e.WriteByte(',')
			}
			e.reflectValue(v.Index(i), false)
		}
		e.WriteByte(']')
	case reflect.Ptr:
		if v.IsNil() {
			e.WriteString("null")
			return
		}
		e.reflectValue(v.Elem(), false)
	default:
		fmt.Println(v, t, t.Kind())
		panic("marshal unsupport type")
	}
}

func marshalerEncoder(e *encodeState, v reflect.Value, quoted bool) {
	if v.Kind() == reflect.Ptr && v.IsNil() {
		e.WriteString("null")
		return
	}
	m := v.Interface().(Marshaler)
	b, err := m.MarshalJSON()
	if err != nil {
		panic(err)
	}
	e.Write(b)
}

func addrMarshalerEncoder(e *encodeState, v reflect.Value, quoted bool) {
	va := v.Addr()
	if va.IsNil() {
		e.WriteString("null")
		return
	}
	m := va.Interface().(Marshaler)
	b, err := m.MarshalJSON()
	if err != nil {
		panic(err)
	}
	e.Write(b)
}
