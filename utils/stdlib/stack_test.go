package stdlib

import (
	"testing"
)

func TestStack(t *testing.T) {
	stack := NewStack[int]()

	// Test IsEmpty on a new stack
	if !stack.IsEmpty() {
		t.Errorf("New stack should be empty")
	}

	// Test Push and Length
	stack.Push(1)
	stack.Push(2)
	if stack.Length() != 2 {
		t.Errorf("Stack length should be 2, got %d", stack.Length())
	}

	// Test Peek
	top := stack.Peek()
	if top != 2 {
		t.Errorf("Top element should be 2, got %d", top)
	}

	// Test Pop
	popped := stack.Pop()
	if popped != 2 {
		t.Errorf("Popped element should be 2, got %d", popped)
	}
	if stack.Length() != 1 {
		t.Errorf("Stack length should be 1 after pop, got %d", stack.Length())
	}

	// Test Pop to empty
	stack.Pop()
	if !stack.IsEmpty() {
		t.Errorf("Stack should be empty after popping all elements")
	}

	// Test Peek on empty stack
	if zeroVal := stack.Peek(); zeroVal != 0 {
		t.Errorf("Peek on empty stack should return zero value, got %d", zeroVal)
	}

	// Test Pop on empty stack
	if zeroVal := stack.Pop(); zeroVal != 0 {
		t.Errorf("Pop on empty stack should return zero value, got %d", zeroVal)
	}
}
