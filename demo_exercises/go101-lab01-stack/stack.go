package stack

// Stack is a LIFO data structure backed by a slice.
// The zero value is an empty, ready-to-use Stack.
type Stack struct {
	items []int
}

// Push adds val to the top of the stack.
func (s *Stack) Push(val int) {
	// TODO: append val to s.items so it becomes the new top element.
	_ = val
}

// Pop removes and returns the top element of the stack.
// It returns (value, true) when the stack is non-empty,
// and (0, false) when the stack is empty.
func (s *Stack) Pop() (int, bool) {
	// TODO: if the stack is empty, return (0, false).
	// TODO: otherwise remove the last element from s.items and return it with true.
	return 0, false
}

// Peek returns the top element without removing it.
// It returns (value, true) when non-empty, (0, false) when empty.
func (s *Stack) Peek() (int, bool) {
	// TODO: if the stack is empty, return (0, false).
	// TODO: otherwise return the last element without modifying s.items.
	return 0, false
}

// IsEmpty reports whether the stack contains no elements.
func (s *Stack) IsEmpty() bool {
	// TODO: report whether there are zero elements in s.items.
	return true
}

// Size returns the number of elements currently in the stack.
func (s *Stack) Size() int {
	// TODO: return the current number of elements in s.items.
	return 0
}
