package sexp

import (
	"fmt"
	"strconv"
	"strings"
)

type ErrNotClosed struct {
	message string
	IsString bool
}

func (e ErrNotClosed) Error() string {
	return e.message
}

// Element 是S表达式的元素
type Element interface {
	// String 返回S表达式的字符串表示
	String() string
}

// Symbol 是S表达式的符号
type Symbol struct {
	Name string
}

// String 返回S表达式的字符串表示
func (s Symbol) String() string {
	return s.Name
}

// String 是S表达式的字符串
type String struct {
	Value string
}

// String 返回S表达式的字符串表示
func (s String) String() string {
	return strconv.Quote(s.Value)
}

// Integer 是S表达式的整数
type Integer int64

// String 返回S表达式的字符串表示
func (i Integer) String() string {
	return fmt.Sprintf("%d", i)
}

// Float 是S表达式的浮点数
type Float float64

// String 返回S表达式的字符串表示
func (f Float) String() string {
	return fmt.Sprintf("%f", f)
}

// List 是S表达式的列表
type List []Element

// String 返回S表达式的字符串表示
func (l List) String() string {
	var parts []string
	for _, e := range l {
		parts = append(parts, e.String())
	}
	return fmt.Sprintf("(%s)", strings.Join(parts, " "))
}

// Parse 将S表达式的字符串表示转换为S表达式
func Parse(file string,s string) (Element, error) {
	// 创建一个新的解析器
	p := parser{
		// 设置文件名
		file: file,
		// 将输入字符串转换为字符切片
		input: []rune(s),
		// 初始化位置
		pos: 0,
	}

	// 调用解析器的parse函数进行解析
	return p.parse()
}

// parser 是S表达式的解析器
type parser struct {
	// 文件名
	file string
	// 输入字符串
	input []rune
	// 当前位置
	pos int
	// 当前行号
	line int
	// 当前列号
	col int
}

// parse 将输入的字符串解析为S表达式
func (p *parser) parse() (Element, error) {
	// 跳过空白字符
	p.skipWhitespace()

	// 获取当前字符
	ch := p.peek()

	// 根据当前字符进行解析
	switch {
	case ch == 0:
		// 如果是EOF，则返回nil
		return nil, nil
	case ch == ')':
		// 如果是右括号，则返回错误
		return nil, fmt.Errorf("unexpected ')' at %s:%d:%d", p.file, p.line, p.col)
	case ch == '(':
		// 如果是左括号，则解析列表
		return p.parseList()
	case ch == '"':
		// 如果是双引号，则解析字符串
		return p.parseString()
	case ch == '+' || ch == '-' || (ch >= '0' && ch <= '9'):
		// 如果是数字，则解析整数或浮点数
		return p.parseNumber()
	default:
		// 否则解析符号
		return p.parseSymbol()
	}
}

// parseList 解析列表
func (p *parser) parseList(ops ...string) (Element, error) {
	// 读取左括号
	p.read()

	// 跳过空白字符
	p.skipWhitespace()

	// 创建一个空列表
	l := List{}

	// 如果当前字符不是右括号，则继续解析列表
	for p.peek() != ')' {
		// 如果是EOF，则返回错误
		if p.peek() == 0 {
			return nil, ErrNotClosed{fmt.Sprintf("expected ')' at %s:%d:%d", p.file, p.line, p.col), false}
		}
		// 解析列表元素
		e, err := p.parse()
		if err != nil {
			return nil, err
		}

		// 将元素添加到列表中
		l = append(l, e)

		// 跳过空白字符
		p.skipWhitespace()
	}

	// 读取右括号
	p.read()

	// 返回列表
	if len(ops) > 0 {
		var op string
		for i:=len(ops)-1; i>=0; i-- {
			op = ops[i]
			switch op {
			case "'":
				l = List{Symbol{"quote"}, l}
			case ",":
				l = List{Symbol{"unquote"}, l}
			default:
				l = List{Symbol{op}, l}
			}
		}
	}
	return l, nil
}

// parseString 解析字符串
func (p *parser) parseString(ops ...string) (Element, error) {

	// 读取左引号
	p.read()

	// 创建一个字符串构建器
	var b strings.Builder

	// 读取字符串
	for {
		// 如果是EOF，则返回错误
		if p.peek() == 0 {
			return nil, ErrNotClosed{fmt.Sprintf("expected '\"' at %s:%d:%d", p.file, p.line, p.col), true}
		}
		// 获取当前字符
		ch := p.read()

		// 如果是右引号，则结束读取
		if ch == '"' {
			break
		}

		// 如果是反斜杠，则读取下一个字符
		if ch == '\\' {
			if len(ops) > 0 && ops[0] == "r" {
				// 如果是原始字符串，则直接将反斜杠添加到字符串构建器中
				b.WriteRune('\\')
				continue
			} else {
				// 如果不是原始字符串，则解析转义字符
				ch, err := p.parseEscape()
				if err != nil {
					return nil, err
				}
				b.WriteRune(ch)
				continue
			}
		}

		// 将字符添加到字符串构建器中
		b.WriteRune(ch)
	}

	// 返回字符串
	return String{b.String()}, nil
}

// parseEscape 解析转义字符
func (p *parser) parseEscape() (rune, error) {
	ch := p.read()
	switch ch {
	case 'n':
		// 换行符
		ch = '\n'
	case 'r':
		// 回车符
		ch = '\r'
	case 't':
		// 制表符
		ch = '\t'
	case 'b':
		// 退格符
		ch = '\b'
	case 'f':
		// 换页符
		ch = '\f'
	case 'e':
		// 颜色转义符
		ch = '\x1b'
	case 'a':
		// 响铃符
		ch = '\a'
	case '\\':
		// 反斜杠
		ch = '\\'
	case '"':
		ch = '"'
	case 'u':
		// 读取0~4个十六进制字符
		var hex string
		for i := 0; i < 4; i++ {
			ch = p.peek()
			if ch >= '0' && ch <= '9' || ch >= 'a' && ch <= 'f' || ch >= 'A' && ch <= 'F' {
				hex += string(p.read())
			} else {
				break
			}
		}

		if len(hex) == 0 {
			return 0, fmt.Errorf("invalid unicode escape character at %s:%d:%d", p.file, p.line, p.col)
		}

		// 将十六进制字符转换为整数
		i, err := strconv.ParseInt(hex, 16, 64)
		if err != nil {
			return 0, err
		}

		// 将整数转换为字符
		ch = rune(i)

	default:
		return 0, fmt.Errorf("invalid escape character: %c at %s:%d:%d", ch, p.file, p.line, p.col)

	}
	return ch, nil
}

// parseNumber 解析数字
func (p *parser) parseNumber() (Element, error) {
	// 创建一个字符串构建器
	var b strings.Builder

	ch := p.peek()
	if ch == '+' || ch == '-' {
		b.WriteRune(p.read())
	}

	// 统计小数点数量
	var hasDot bool

	// 读取数字
	for {
		// 获取当前字符
		ch := p.peek()

		if ch == '.' {
			if hasDot {
				return nil, fmt.Errorf("invalid number at %s:%d:%d", p.file, p.line, p.col)
			}
			hasDot = true
		}

		// 如果是数字或者小数点，则读取字符
		if (ch >= '0' && ch <= '9') || ch == '.' {
			b.WriteRune(p.read())
		} else {
			break
		}
	}

	// 将字符串转换为数字
	if hasDot {
		f, err := strconv.ParseFloat(b.String(), 64)
		if err != nil {
			return nil, err
		}
		return Float(f), nil
	} else {
		i, err := strconv.ParseInt(b.String(), 10, 64)
		if err != nil {
			return nil, err
		}
		return Integer(i), nil
	}
}

func splitOp(opString string) []string {
	var ops []string
	var b strings.Builder
	for _, ch := range opString {
		if ch == '\'' || ch == ',' {
			if b.Len() > 0 {
				ops = append(ops, b.String())
				b.Reset()
			}
			ops = append(ops, string(ch))
		} else {
			b.WriteRune(ch)
		}
	}
	if b.Len() > 0 {
		ops = append(ops, b.String())
	}
	return ops
}

// parseSymbol 解析符号
func (p *parser) parseSymbol() (Element, error) {
	// 创建一个字符串构建器
	var b strings.Builder

	// 读取符号
	for {
		// 获取当前字符
		ch := p.peek()

		// 如果是空白字符或者右括号，则结束读取
		if ch == 0 || ch == ' ' || ch == '\t' || ch == '\r' || ch == '\n' || ch == ')' {
			break
		}

		// 如果是左括号，则解析列表
		if ch == '(' {
			return p.parseList(splitOp(b.String())...)
		}
		// 如果是双引号
		if ch == '"' {
			return p.parseString(b.String())
		}

		// 读取字符
		b.WriteRune(p.read())
	}

	// 返回符号
	return Symbol{b.String()}, nil
}

// skipWhitespace 跳过空白字符
func (p *parser) skipWhitespace() {
	for {
		// 获取当前字符
		ch := p.peek()

		// 如果是空白字符，则继续读取
		if ch == ' ' || ch == '\t' || ch == '\r' || ch == '\n' {
			p.read()
		} else {
			break
		}
	}
}

// peek 返回当前字符
func (p *parser) peek() rune {
	if p.pos >= len(p.input) {
		return 0
	}
	return p.input[p.pos]
}

// read 读取当前字符，并将位置后移，同时统计行号和列号
func (p *parser) read() rune {
	// fmt.Printf("read %d: %c", p.pos, p.peek())
	ch := p.peek()
	if ch == '\n' {
		p.line++
		p.col = 0
	} else {
		p.col++
	}
	p.pos++
	return ch
}


