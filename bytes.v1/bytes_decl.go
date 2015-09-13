package bytes

//go:noescape
//这是什么特殊的标记?

// 汇编实现runtime/asm_$GOARCH.s
// IndexByte会用c在s里查找, 如果找到了则返回位置index, 找不到则返回-1, 用汇编实现,效率会十分高的
func IndexByte(s []byte, c byte) int

// Equal 这里有实现, 为何在还有一个equalProtable函数?
// 这里有一个特殊: 一个nil参数与任何空切片相等
func Equal(a, b []byte) bool

// Compare ...
func Compare(a, b []byte) int
