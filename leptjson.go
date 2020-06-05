package goleptjson

import (
	"bytes"
	"encoding/binary"
	"errors"
	"math"
	"strconv"
)

var (
	// ErrReachEnd the input string is reach to end
	ErrReachEnd = errors.New("json string reach end")
	// ErrUnexpectChar expect function fail
	ErrUnexpectChar = errors.New("get an unexpect char")
)

// define some global parse events
const (
	// LeptParseOK just ok
	LeptParseOK int = iota
	// LeptParseExpectValue expect value
	LeptParseExpectValue
	// LeptParseInvalidValue invalid value
	LeptParseInvalidValue
	// LeptParseRootNotSingular root not singular
	LeptParseRootNotSingular
	// LeptParseNumberTooBig number is to big
	LeptParseNumberTooBig
	// LeptParseMissQuotationMark
	LeptParseMissQuotationMark
	// LeptParseInvalidStringEscape
	LeptParseInvalidStringEscape
	// LeptParseInvalidStringChar
	LeptParseInvalidStringChar
	// LeptParseInvalidUnicodeHex
	LeptParseInvalidUnicodeHex
	// LeptParseInvalidUnicodeSurrogate
	LeptParseInvalidUnicodeSurrogate
	// LeptParseMissCommaOrSouareBracket
	LeptParseMissCommaOrSouareBracket
)

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

// LeptValue hold the value
type LeptValue struct {
	typ LeptType
	n   float64
	s   string
	e   []*LeptValue // for array
}

// NewLeptValue return a init LeptValue
func NewLeptValue() *LeptValue {
	return &LeptValue{
		typ: LeptFalse,
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
func LeptParseNull(c *LeptContext, v *LeptValue) int {
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
func LeptParseTrue(c *LeptContext, v *LeptValue) int {
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
func LeptParseFalse(c *LeptContext, v *LeptValue) int {
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
func LeptParseLiteral(c *LeptContext, v *LeptValue, literal string, typ LeptType) int {
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
func LeptParseNumber(c *LeptContext, v *LeptValue) int {
	var end string
	var err error
	v.n, end, err = strtod(c.json)
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
	if input[0] == '0' && n > 1 && !(input[1] == '.' || input[1] == 'e' || input[1] == 'E') && isDigit(input[1]) {
		// start with zero illegal like 0123
		return ret, input, IllegalInput
	}
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
		ret += float64(decimal) / float64(frac)
		if len(input) == 0 {
			if neg {
				return -ret, input, nil
			}
			return ret, input, nil
		}
		if !(input[0] == 'e' || input[0] == 'E') {
			// following is not exp
			return ret, input, IllegalInput
		}
		input, exp, err = parseExp(input)
		if err != nil || len(input) != 0 {
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
		// get exp
		input, exp, err = parseExp(input)
		if err != nil {
			return ret, input, err
		}
		if len(input) != 0 {
			// follow illegal char
			return ret, input, IllegalInput
		}
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
func LeptParseString(c *LeptContext, v *LeptValue) int {
	expect(c, '"')
	var stack bytes.Buffer
	for i, n := 0, len(c.json); i < n; i++ {
		ch := c.json[i]
		switch ch {
		case '"':
			LeptSetString(v, stack.String())
			stack.Truncate(0)
			c.json = c.json[i+1:]
			return LeptParseOK
		case '\\':
			// 遇到第一个转义符号，需要连续匹配两个 \
			if i+1 >= n {
				return LeptParseInvalidStringEscape
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
			case 'u':
				u, err := leptParseHex4(c.json[i+2:])
				if err != nil {
					return LeptParseInvalidUnicodeHex
				}
				if u < 0 || u > 0x10FFFF {
					return LeptParseInvalidUnicodeHex
				}
				if u >= 0xD800 && u <= 0xDBFF { /* surrogate pair */
					if i+6 >= n || c.json[i+6] != '\\' {
						return LeptParseInvalidUnicodeSurrogate
					}
					if i+7 >= n || c.json[i+7] != 'u' {
						return LeptParseInvalidUnicodeSurrogate
					}
					u2, err := leptParseHex4(c.json[i+8:])
					if err != nil {
						return LeptParseInvalidUnicodeHex
					}
					if u2 < 0xDC00 || u2 > 0xDFFF {
						return LeptParseInvalidUnicodeSurrogate
					}
					u = (((u - 0xD800) << 10) | (u2 - 0xDC00)) + 0x10000
					i += 6
				}
				// 检查代理对
				// stack.WriteString(leptEncodeUTF8(u))
				if u <= 0x7F {
					stack.Write(leptEncodeUTF8(u & 0xFF))
				} else if u <= 0x7FF {
					stack.Write(leptEncodeUTF8(0xC0 | ((u >> 6) & 0xFF)))
					stack.Write(leptEncodeUTF8(0x80 | (u & 0x3F)))
				} else if u <= 0xFFFF {
					stack.Write(leptEncodeUTF8(0xE0 | ((u >> 12) & 0xFF)))
					stack.Write(leptEncodeUTF8(0x80 | ((u >> 6) & 0x3F)))
					stack.Write(leptEncodeUTF8(0x80 | (u & 0x3F)))
				} else if u <= 0x10FFFF {
					stack.Write(leptEncodeUTF8(0xF0 | ((u >> 18) & 0xFF)))
					stack.Write(leptEncodeUTF8(0x80 | ((u >> 12) & 0x3F)))
					stack.Write(leptEncodeUTF8(0x80 | ((u >> 6) & 0x3F)))
					stack.Write(leptEncodeUTF8(0x80 | (u & 0x3F)))
				} else {
					panic("u is illegal")
				}
				// 将 uxxxx 跳过
				// \\ 最后是有 i++ 这里只需要 4
				i += 4
			default:
				return LeptParseInvalidStringEscape
			}
			// 这里的 i++ 针对普通的转码字符，至于 unicode 需要另外处理 uxxxx 个字符
			i++
		default:
			// 	unescaped = %x20-21 / %x23-5B / %x5D-10FFFF
			// 当中空缺的 %x22 是双引号，%x5C 是反斜线，都已经处理。所以不合法的字符是 %x00 至 %x1F。
			if ch < 0x20 {
				return LeptParseInvalidStringChar
			}
			stack.WriteByte(ch)
		}
	}
	// reach end of string becase the string has no \"
	return LeptParseMissQuotationMark
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

// func hex() {
// 	bufSize := 8
// 	buf := make([]byte, bufSize)
// 	write := 0
// 	if u <= 0x007F {
// 		write = binary.PutUvarint(buf, u&0xFF)
// 		stack.Write(buf[:write])
// 	} else if u >= 0x0080 && u <= 0x07FF {
// 		write = binary.PutUvarint(buf, 0xC0|((u>>6)&0xFF))
// 		stack.Write(buf[:write])
// 		write = binary.PutUvarint(buf, 0x80|(u&0x3F))
// 		stack.Write(buf[:write])
// 	} else if u >= 0x0800 && u <= 0xFFFF {
// 		write = binary.PutUvarint(buf, 0xE0|((u>>12)&0xFF))
// 		stack.Write(buf[:write])
// 		write = binary.PutUvarint(buf, 0x80|((u>>6)&0x3F))
// 		stack.Write(buf[:write])
// 		write = binary.PutUvarint(buf, 0x80|(u&0x3F))
// 		stack.Write(buf[:write])
// 	} else if u >= 0x10000 && u <= 0x10FFFF {
// 		write = binary.PutUvarint(buf, 0xF0|((u>>18)&0xFF))
// 		stack.Write(buf[:write])
// 		write = binary.PutUvarint(buf, 0x80|((u>>12)&0x3F))
// 		stack.Write(buf[:write])
// 		write = binary.PutUvarint(buf, 0x80|((u>>6)&0x3F))
// 		stack.Write(buf[:write])
// 		write = binary.PutUvarint(buf, 0x80|(u&0x3F))
// 		stack.Write(buf[:write])
// 	} else {
// 		panic("u is illegal")
// 	}
// }
// func leptEncodeUTF8(u uint64) string {
// 	// 针对 四个区间         码点位数   字节1      字节2      字节3     字节4
// 	// 0x0000 - 0x007F      7         0xxxxxxx
// 	// 0x0080 - 0x07FF      11        1100xxxx   10xxxxxx
// 	// 0x0800 - 0xFFFF      16        1110xxxx   10xxxxxx  10xxxxxx
// 	// 0x10000 - 0x10FFFF   21        11110xxx   10xxxxxx  10xxxxxx  10xxxxxx
// 	if u <= 0x007F {
// 		return formatUintToHex(u)
// 	}
// 	if u >= 0x0080 && u <= 0x07FF {
// 		return formatUintToHex(0xC0|((u>>6)&0xFF)) +
// 			formatUintToHex(0x80|(u&0x3F))
// 	}
// 	if u >= 0x0800 && u <= 0xFFFF {
// 		return formatUintToHex(0xE0|((u>>12)&0xFF)) +
// 			formatUintToHex(0x80|((u>>6)&0x3F)) +
// 			formatUintToHex(0x80|(u&0x3F))
// 	}
// 	if u >= 0x10000 && u <= 0x10FFFF {
// 		return formatUintToHex(0xF0|((u>>18)&0xFF)) +
// 			formatUintToHex(0x80|((u>>12)&0x3F)) +
// 			formatUintToHex(0x80|((u>>6)&0x3F)) +
// 			formatUintToHex(0x80|(u&0x3F))
// 	}
// 	return "illegal-utf8-string"
// }
// func formatUintToHex(num uint64) string {
// 	return strconv.FormatUint(num, 16)
// }

// LeptParseValue use to parse value switch to spec func
func LeptParseValue(c *LeptContext, v *LeptValue) int {
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
	default:
		return LeptParseNumber(c, v)
	}
}

// LeptParseArray use to parse array
func LeptParseArray(c *LeptContext, v *LeptValue) int {
	expect(c, '[')
	LeptParseWhitespace(c)
	n := len(c.json)
	if n == 0 {
		return LeptParseMissCommaOrSouareBracket
	}
	if c.json[0] == ']' {
		v.typ = LeptArray
		v.e = make([]*LeptValue, 0)
		c.json = c.json[1:]
		return LeptParseOK
	}
	for {
		LeptParseWhitespace(c) // my
		vi := NewLeptValue()
		if ok := LeptParseValue(c, vi); ok != LeptParseOK {
			return ok
		}
		v.e = append(v.e, vi)
		LeptParseWhitespace(c) //my

		// LeptParseWhitespace(c) // tutorial
		if len(c.json) == 0 {
			return LeptParseMissCommaOrSouareBracket
		}
		if c.json[0] == ',' {
			c.json = c.json[1:]
			// LeptParseWhitespace(c) // tutorial
		} else if c.json[0] == ']' {
			c.json = c.json[1:]
			v.typ = LeptArray
			return LeptParseOK
		} else {
			return LeptParseMissCommaOrSouareBracket
		}
	}
}

// LeptParse use to parse value the enter
func LeptParse(v *LeptValue, json string) int {
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
	if len(v.e) <= index {
		panic("LeptGetArrayElement v length <= input index")
	}
	return v.e[index]
}

// LeptGetArraySize use to get the size of array
func LeptGetArraySize(v *LeptValue) int {
	if v == nil || v.typ != LeptArray {
		panic("LeptGetArrayElement v is nil or typ is not array")
	}
	return len(v.e)
}
