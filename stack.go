package main

type Element struct {
	Row, Col int
	Char     rune
}

type Stack []Element

func NewStack() Stack {
	return make(Stack, 0)
}

func (s Stack) Push(e Element) Stack {
	return append(s, e)
}

func (s Stack) Pop() (Stack, Element) {
	l := len(s)
	if l == 0 {
		return s, Element{
			Row: -1,
			Col: -1,
		}
	}
	return s[:l-1], s[l-1]
}
