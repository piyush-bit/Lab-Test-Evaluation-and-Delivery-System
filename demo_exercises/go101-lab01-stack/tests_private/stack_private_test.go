package stack

import (
	"fmt"
	"testing"
)

func passPrivate(t *testing.T, name string) {
	t.Helper()
	fmt.Printf("[PASS] %s\n", name)
}

func TestPrivate_LIFOOrder(t *testing.T) {
	var s Stack
	for i := 1; i <= 5; i++ {
		s.Push(i)
	}
	for i := 5; i >= 1; i-- {
		val, ok := s.Pop()
		if !ok || val != i {
			t.Fatalf("expected (%d, true) on pop, got (%d, %v)", i, val, ok)
		}
	}
	if !s.IsEmpty() {
		t.Fatal("stack must be empty after popping all elements")
	}
	passPrivate(t, "Private 1: LIFO ordering across 5 elements")
}

func TestPrivate_InterleavedOps(t *testing.T) {
	var s Stack
	s.Push(1)
	s.Push(2)
	s.Pop()         // removes 2
	s.Push(3)       // stack: [1, 3]
	val, ok := s.Peek()
	if !ok || val != 3 {
		t.Fatalf("expected Peek == 3 after interleaved ops, got (%d, %v)", val, ok)
	}
	if s.Size() != 2 {
		t.Fatalf("expected Size == 2, got %d", s.Size())
	}
	passPrivate(t, "Private 2: Interleaved Push / Pop / Peek")
}

func TestPrivate_SizeUnderLoad(t *testing.T) {
	var s Stack
	const n = 1000
	for i := 0; i < n; i++ {
		s.Push(i)
	}
	if s.Size() != n {
		t.Fatalf("expected Size == %d, got %d", n, s.Size())
	}
	for i := 0; i < n/2; i++ {
		s.Pop()
	}
	if s.Size() != n/2 {
		t.Fatalf("expected Size == %d after %d pops, got %d", n/2, n/2, s.Size())
	}
	passPrivate(t, "Private 3: Size tracking under load (1000 pushes, 500 pops)")
}

func TestPrivate_PopUntilEmpty(t *testing.T) {
	var s Stack
	s.Push(7)
	s.Push(8)
	s.Pop()
	s.Pop()
	_, ok := s.Pop() // one extra pop on now-empty stack
	if ok {
		t.Fatal("Pop on exhausted stack must return ok == false")
	}
	if !s.IsEmpty() {
		t.Fatal("stack must report IsEmpty after all elements removed")
	}
	passPrivate(t, "Private 4: Pop past empty does not corrupt state")
}

func TestPrivate_PeekPreservesOrder(t *testing.T) {
	var s Stack
	s.Push(100)
	s.Push(200)
	for i := 0; i < 5; i++ {
		val, ok := s.Peek()
		if !ok || val != 200 {
			t.Fatalf("Peek call %d: expected (200, true), got (%d, %v)", i+1, val, ok)
		}
	}
	if s.Size() != 2 {
		t.Fatal("repeated Peek calls must not change the stack size")
	}
	passPrivate(t, "Private 5: Repeated Peek does not mutate stack")
}
