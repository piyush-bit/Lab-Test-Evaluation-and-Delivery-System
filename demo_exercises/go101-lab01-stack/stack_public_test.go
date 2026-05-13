package stack

import (
	"fmt"
	"testing"
)

func passPublic(t *testing.T, name string) {
	t.Helper()
	fmt.Printf("[PASS] %s\n", name)
}

func TestPushAndSize(t *testing.T) {
	var s Stack
	if !s.IsEmpty() {
		t.Fatal("new stack should be empty")
	}
	s.Push(1)
	s.Push(2)
	s.Push(3)
	if s.Size() != 3 {
		t.Fatalf("expected Size() == 3 after three pushes, got %d", s.Size())
	}
	if s.IsEmpty() {
		t.Fatal("stack should not be empty after pushes")
	}
	passPublic(t, "Task 1: Push and Size")
}

func TestPopBasic(t *testing.T) {
	var s Stack
	s.Push(10)
	s.Push(20)
	val, ok := s.Pop()
	if !ok {
		t.Fatal("Pop on non-empty stack must return ok == true")
	}
	if val != 20 {
		t.Fatalf("expected top element 20 (LIFO), got %d", val)
	}
	if s.Size() != 1 {
		t.Fatalf("expected Size() == 1 after one pop, got %d", s.Size())
	}
	passPublic(t, "Task 2: Pop returns top element (LIFO)")
}

func TestPopOnEmpty(t *testing.T) {
	var s Stack
	val, ok := s.Pop()
	if ok {
		t.Fatal("Pop on empty stack must return ok == false")
	}
	if val != 0 {
		t.Fatalf("Pop on empty stack must return value 0, got %d", val)
	}
	passPublic(t, "Task 3: Pop on empty stack returns (0, false)")
}

func TestPeekDoesNotRemove(t *testing.T) {
	var s Stack
	s.Push(42)
	val, ok := s.Peek()
	if !ok {
		t.Fatal("Peek on non-empty stack must return ok == true")
	}
	if val != 42 {
		t.Fatalf("expected Peek value 42, got %d", val)
	}
	if s.Size() != 1 {
		t.Fatal("Peek must not remove the element; Size should still be 1")
	}
	passPublic(t, "Task 4: Peek does not remove the top element")
}

func TestPeekOnEmpty(t *testing.T) {
	var s Stack
	_, ok := s.Peek()
	if ok {
		t.Fatal("Peek on empty stack must return ok == false")
	}
	passPublic(t, "Task 5: Peek on empty stack returns (0, false)")
}
