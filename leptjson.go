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
		input, exp, err = parseExp(input)
		if err != nil {
			return ret, input, err
		}
		if len(input) != 0 {
			// do not parse any more leave it to next parser
			return ret, input, nil
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
			case 'u':
				u, err := leptParseHex4(c.json[i+2:])
				if err != nil {
					return "", LeptParseInvalidUnicodeHex
				}
				if u < 0 || u > 0x10FFFF {
					return "", LeptParseInvalidUnicodeHex
				}
				if u >= 0xD800 && u <= 0xDBFF { /* surrogate pair */
					if i+6 >= n || c.json[i+6] != '\\' {
						return "", LeptParseInvalidUnicodeSurrogate
					}
					if i+7 >= n || c.json[i+7] != 'u' {
						return "", LeptParseInvalidUnicodeSurrogate
					}
					u2, err := leptParseHex4(c.json[i+8:])
					if err != nil {
						return "", LeptParseInvalidUnicodeHex
					}
					if u2 < 0xDC00 || u2 > 0xDFFF {
						return "", LeptParseInvalidUnicodeSurrogate
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
		if len(ki) == 0 {
			return LeptParseMissKey
		}
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
		return strconv.FormatFloat(v.n, 'g', -1, 64)
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
