package goleptjson

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

const jsonchecker string = "./data/jsonchecker"
const roundtrip string = "./data/roundtrip"

// pathExists use os.stat to check path
func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
func readJSON(path string) (string, error) {
	exists, err := pathExists(path)
	if err != nil || !exists {
		return "", err
	}
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

// 这里复用 leptjson_test.go 的断言函数
func TestFailJSON(t *testing.T) {
	for i := 1; i <= 31; i++ {
		path := filepath.Join(jsonchecker, fmt.Sprintf("fail%02d.json", i))
		buf, err := readJSON(path)
		if err != nil {
			t.Errorf("readJSON %v get err: %v", path, err)
		}
		if buf == "" {
			// skip the exclue.json
			continue
		}
		v := NewLeptValue()
		expectEQBool(t, true, LeptParse(v, buf) != LeptParseOK)
	}
}
func TestPassJSON(t *testing.T) {
	for i := 1; i <= 3; i++ {
		path := filepath.Join(jsonchecker, fmt.Sprintf("pass%02d.json", i))
		buf, err := readJSON(path)
		if err != nil {
			t.Errorf("readJSON %v get err: %v", path, err)
		}
		if buf == "" {
			// skip the exclue.json
			continue
		}
		v := NewLeptValue()
		event := LeptParse(v, buf)
		expectEQBool(t, true, event == LeptParseOK)
	}
}
func TestRoundtripJSON(t *testing.T) {
	for i := 1; i <= 27; i++ {
		path := filepath.Join(roundtrip, fmt.Sprintf("roundtrip%02d.json", i))
		buf, err := readJSON(path)
		if err != nil {
			t.Errorf("readJSON %v get err: %v", path, err)
		}
		if buf == "" {
			// skip the exclue.json
			continue
		}
		v := NewLeptValue()
		event := LeptParse(v, buf)
		expectEQBool(t, true, event == LeptParseOK)
		// actual := LeptStringify(v)
		// 对于大整数的处理不够灵活，可以考虑区分没有溢出的整数和浮点数
		// expectEQString(t, buf, actual)
	}
}

func BenchmarkCanadaJSON(b *testing.B) {
	path := filepath.Join("./data", "canada.json")
	buf, err := readJSON(path)
	if err != nil {
		b.Errorf("readJSON %v get err: %v", path, err)
		return
	}
	if buf == "" {
		b.Errorf("readJSON %v get empty string : %v", path, err)
		return
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v := NewLeptValue()
		event := LeptParse(v, buf)
		if event != LeptParseOK {
			b.Errorf("benchmark parse err : %v", event)
		}
		// fmt.Println(event)
	}
}
func BenchmarkCitmCatalogJSON(b *testing.B) {
	path := filepath.Join("./data", "citm_catalog.json")
	buf, err := readJSON(path)
	if err != nil {
		b.Errorf("readJSON %v get err: %v", path, err)
		return
	}
	if buf == "" {
		b.Errorf("readJSON %v get empty string : %v", path, err)
		return
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v := NewLeptValue()
		event := LeptParse(v, buf)
		if event != LeptParseOK {
			b.Errorf("benchmark parse err : %v", event)
		}
		// fmt.Println(event)
	}
}
func BenchmarkTwitterJSON(b *testing.B) {
	path := filepath.Join("./data", "twitter.json")
	buf, err := readJSON(path)
	if err != nil {
		b.Errorf("readJSON %v get err: %v", path, err)
		return
	}
	if buf == "" {
		b.Errorf("readJSON %v get empty string : %v", path, err)
		return
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v := NewLeptValue()
		event := LeptParse(v, buf)
		if event != LeptParseOK {
			b.Errorf("benchmark parse err : %v", event)
		}
		// fmt.Println(event)
	}
}
