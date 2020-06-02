package leptjson

import (
	"errors"
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
)

// LeptType enums of json type
type LeptType int

const (
	// LeptNULL parse to nil
	LeptNULL LeptType = iota
	// LeptFALSE parse to false
	LeptFALSE
	// LeptTRUE parse to true
	LeptTRUE
	// LeptNUMBER parse to number like int float
	LeptNUMBER
	// LeptSTRING parse to string
	LeptSTRING
	// LeptARRAY parse to array
	LeptARRAY
	// LeptOBJECT parse to map
	LeptOBJECT
)

// LeptValue hold the value
type LeptValue struct {
	typ LeptType
}

// LeptContext hold the input string
type LeptContext struct {
	json string
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
	v.typ = LeptNULL
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
	v.typ = LeptTRUE
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
	v.typ = LeptFALSE
	return LeptParseOK
}

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
	default:
		return LeptParseInvalidValue
	}
}

// LeptParse use to parse value the enter
func LeptParse(v *LeptValue, json string) int {
	if v == nil {
		panic("LeptParse v is nil")
	}
	c := &LeptContext{json}
	v.typ = LeptNULL
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
