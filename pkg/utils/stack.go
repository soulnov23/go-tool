package utils

type Stack []any

func (s *Stack) Push(value any) {
	*s = append(*s, value)
}

func (s *Stack) Pop() any {
	if size := len(*s); size > 0 {
		value := (*s)[size-1]
		*s = (*s)[:size-1]
		return value
	}
	return nil
}
