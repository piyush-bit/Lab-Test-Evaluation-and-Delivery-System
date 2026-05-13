package stack

// Stack is a LIFO data structure backed by a slice.
// The zero value is an empty, ready-to-use Stack.
type Stack struct {
	items []int
}

// Push adds val to the top of the stack.
func (s *Stack) Push(val int) {
	s.items = append(s.items, val)
}

// Pop removes and returns the top element of the stack.
// It returns (value, true) when the stack is non-empty,
// and (0, false) when the stack is empty.
func (s *Stack) Pop() (int, bool) {
	if len(s.items) == 0 {
		return 0, false
	}
	top := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return top, true
}

// Peek returns the top element without removing it.
// It returns (value, true) when non-empty, (0, false) when empty.
func (s *Stack) Peek() (int, bool) {
	if len(s.items) == 0 {
		return 0, false
	}
	return s.items[len(s.items)-1], true
}

// IsEmpty reports whether the stack contains no elements.
func (s *Stack) IsEmpty() bool {
	return len(s.items) == 0
}

// Size returns the number of elements currently in the stack.
func (s *Stack) Size() int {
	return len(s.items)
}
