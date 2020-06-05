## goleptjson

just a json parser learning resp!

main idea and lesson comes from (miloyip/json-tutorial)[https://github.com/miloyip/json-tutorial]


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
unicode
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


