## goleptjson

just a json parser learning resp!

main idea and lesson comes from (miloyip/json-tutorial)[https://github.com/miloyip/json-tutorial]

### Parse Error
leptjson 使用 int 作为解析错误的返回值，这里是否需要修改
1.自定义 error 类型，使用 go 的 err != nil 方式
2.使用 int 的一个类型，结构保持一致， ErrorType == 0 || 1 || 2 || 3

### number
```md
	// number = [ "-" ] int [ frac ] [ exp ]
	// int = "0" / digit1-9 *digit
	// frac = "." 1*digit
	// exp = ("e" / "E") ["-" / "+"] 1*digit
```
### string
```md
string = quotation-mark *char quotation-mark
char = unescaped /
   escape (
       %x22 /          ; "    quotation mark  U+0022
       %x5C /          ; \    reverse solidus U+005C
       %x2F /          ; /    solidus         U+002F
       %x62 /          ; b    backspace       U+0008
       %x66 /          ; f    form feed       U+000C
       %x6E /          ; n    line feed       U+000A
       %x72 /          ; r    carriage return U+000D
       %x74 /          ; t    tab             U+0009
       %x75 4HEXDIG )  ; uXXXX                U+XXXX
escape = %x5C          ; \
quotation-mark = %x22  ; "
unescaped = %x20-21 / %x23-5B / %x5D-10FFFF
```
### unicode
```md
我们举一个例子解析多字节的情况，欧元符号 € → U+20AC：

U+20AC 在 U+0800 ~ U+FFFF 的范围内，应编码成 3 个字节。
U+20AC 的二进位为 10000010101100
3 个字节的情况我们要 16 位的码点，所以在前面补两个 0，成为 0010000010101100
按上表把二进位分成 3 组：0010, 000010, 101100
加上每个字节的前缀：11100010, 10000010, 10101100
用十六进位表示即：0xE2, 0x82, 0xAC
对于这例子的范围，对应的 C 代码是这样的：

此时 u 应该为 10000010101100
u >> 12 得到 0010 
u >> 12 & 0xFF = 0010 & 1111 1111 = 0000 0010
0xE0 | ((u >> 12) & 0xFF) = 1110 0000(也就是对应表格的码点字节头) | 0000 0010 = 1110 0010 = 0xE2
下面几个位的类似
if (u >= 0x0800 && u <= 0xFFFF) {
    OutputByte(0xE0 | ((u >> 12) & 0xFF)); /* 0xE0 = 11100000 */

    OutputByte(0x80 | ((u >>  6) & 0x3F)); /* 0x80 = 10000000 */
    OutputByte(0x80 | ( u        & 0x3F)); /* 0x3F = 00111111 */
}
```
未解决 encode uint64 的问题
代理对的解析出现了错误
~~fix 对于 uint64 转为输出的 hex 字符串需要如何处理，现在是使用 []byte 结合 buffer 生成，
也就是关于 go 这些格式数据的转化问题不清晰明了

对于 uxxxx 的 utf8 编码字符串，需要考虑代理对的问题，
对于 u >= 0xD800 && u <= 0xDBFF
u = (((u - 0xD800) << 10) | (u2 - 0xDC00)) + 0x10000
更新对应的字符串的值。
而对于 u 可以区分四个区间的值，参考
```go
// 	// 针对 四个区间         码点位数   字节1      字节2      字节3     字节4
// 	// 0x0000 - 0x007F      7         0xxxxxxx
// 	// 0x0080 - 0x07FF      11        1100xxxx   10xxxxxx
// 	// 0x0800 - 0xFFFF      16        1110xxxx   10xxxxxx  10xxxxxx
// 	// 0x10000 - 0x10FFFF   21        11110xxx   10xxxxxx  10xxxxxx  10xxxxxx

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
```
所以对于 u，针对每一个区间计算出对应的 uint64，再使用 leptEncodeUTF8 得到可以写入 buffer 的 []byte 数组
buffer.String() 可以得到完整的 utf8 解码字符串

### array
```md
array = %x5B ws [ value *( ws %x2C ws value ) ] ws %x5D
当中，%x5B 是左中括号 [，%x2C 是逗号 ,，%x5D 是右中括号 ] ，ws 是空白字符。一个数组可以包含零至多个值，以逗号分隔，例如 []、[1,2,true]、[[1,2],[3,4],"abc"] 都是合法的数组。但注意 JSON 不接受末端额外的逗号，例如 [1,2,] 是不合法的（许多编程语言如 C/C++、Javascript、Java、C# 都容许数组初始值包含末端逗号）。
````


### object
```md
JSON 对象和 JSON 数组非常相似，区别包括 JSON 对象以花括号 {}（U+007B、U+007D）包裹表示，另外 JSON 对象由对象成员（member）组成，而 JSON 数组由 JSON 值组成。所谓对象成员，就是键值对，键必须为 JSON 字符串，然后值是任何 JSON 值，中间以冒号 :（U+003A）分隔。完整语法如下：

member = string ws %x3A ws value
object = %x7B ws [ member *( ws %x2C ws member ) ] ws %x7D
```
### array object
两者大体解析过程是相似的，不过存放的地址不同，object 多了解析 key 值得步骤。
这里都是使用 slice 存储具体值，可以针对 object 优化实现哈希链表的结构，更加高效。

### interface{}
golang 提供的对象是 interface{} 可以存储 nil,bool,number,string,slice,map 
提供三个方法解析 LeptValue
```go
func ToInterface(v *LeptValue) interface{} {
}
func ToMap(v *LeptValue) map[string]interface{} {
}
func ToArray(v *LeptValue) []interface{} {
}
```
在对应的基础上，可以使用 []struct{} struct{} 实现 encoding/json 中的
方法，将 json 字符串映射到 struct 中。
```go
func ToStruct(v *LeptValue, structure interface{}) error {
}
// ToStruct(v, &struct{a int, b string}{})
// ToStruct(v, &[]struct{a int, b string}{})
```
映射为 struct 时，以传入的 struct 为参考，如果 v 的类型或者值不对应的话，会返回错误。
对于初始化的值，不知道是否有默认值，现时，struct 的全部字段都会设置默认值。
```go
{<nil> false true 123 abc [] map[]}
```
未知原因导致 数组和 map 解析不正确
数组需要初始化为一个合适的 cap 的 slice
map 需要知道 key value 的 Type
方法：
reflect.TypeOf(m).Key()
reflect.TypeOf(m).Elem()
```GO
// 生成的 rvt 是 reflect.flag.mustBeAssignable using unaddressable value
// 导致无法在 toMap 之后得到 rvt 的值。
// false false string map[string]interface {} true map[]
rvt := reflect.MakeMapWithSize(reflect.MapOf(reflect.TypeOf("abc"), rv.Type()), len(v.o))
fmt.Println(rvt.CanAddr(), rvt.CanSet(), reflect.TypeOf("abc"), reflect.MapOf(reflect.TypeOf("abc"), rv.Type()), rvt.CanInterface(), rvt)
toMap(v, rvt)
rv.Set(rvt)
```
