package stdlib

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestLinkedList_AddAndCount(t *testing.T) {
	ll := NewLinkedList[int]()
	if ll.Count() != 0 {
		t.Errorf("Expected count 0, got %d", ll.Count())
	}

	ll.AddAtBeg(10)
	ll.AddAtBeg(20)
	ll.AddAtEnd(30)

	if ll.Count() != 3 {
		t.Errorf("Expected count 3, got %d", ll.Count())
	}
}

func TestLinkedList_DelAtBeg(t *testing.T) {
	ll := NewLinkedList[int]()
	_, ok := ll.DelAtBeg()
	if ok {
		t.Error("Expected false on empty list")
	}

	ll.AddAtBeg(10)
	val, ok := ll.DelAtBeg()
	if !ok || val != 10 {
		t.Errorf("Expected 10, true; got %d, %v", val, ok)
	}
}

func TestLinkedList_DelAtEnd(t *testing.T) {
	ll := NewLinkedList[int]()
	_, ok := ll.DelAtEnd()
	if ok {
		t.Error("Expected false on empty list")
	}

	ll.AddAtBeg(10)
	val, ok := ll.DelAtEnd()
	if !ok || val != 10 {
		t.Errorf("Expected 10, true; got %d, %v", val, ok)
	}

	ll.AddAtBeg(20)
	ll.AddAtEnd(30)
	val, ok = ll.DelAtEnd()
	if !ok || val != 30 {
		t.Errorf("Expected 30, true; got %d, %v", val, ok)
	}
}

func TestLinkedList_DelByPos(t *testing.T) {
	ll := NewLinkedList[int]()
	ll.AddAtEnd(10) // position 1
	ll.AddAtEnd(20) // position 2
	ll.AddAtEnd(30) // position 3

	val, ok := ll.DelByPos(2)
	if !ok || val != 20 {
		t.Errorf("Expected 20, true; got %d, %v", val, ok)
	}

	if ll.Count() != 2 {
		t.Errorf("Expected length 2 after deletion, got %d", ll.Count())
	}

	val, ok = ll.DelByPos(1)
	if !ok || val != 10 {
		t.Errorf("Expected 10, true; got %d, %v", val, ok)
	}

	val, ok = ll.DelByPos(1)
	if !ok || val != 30 {
		t.Errorf("Expected 30, true; got %d, %v", val, ok)
	}

	val, ok = ll.DelByPos(1)
	if ok {
		t.Errorf("Expected false for deletion from empty list, got %v with val %d", ok, val)
	}
}

func TestLinkedList_DelByPos_EdgeCases(t *testing.T) {
	ll := NewLinkedList[int]()
	ll.AddAtEnd(1)
	ll.AddAtEnd(2)

	// pos-1 == Count triggers DelAtEnd
	val, ok := ll.DelByPos(3)
	if !ok || val != 2 {
		t.Errorf("Expected 2, true; got %d, %v", val, ok)
	}

	// Out-of-bounds should return false
	val, ok = ll.DelByPos(10)
	if ok {
		t.Errorf("Expected false for out of bounds deletion, got %v with val %d", ok, val)
	}
}

func TestLinkedList_Reverse(t *testing.T) {
	ll := NewLinkedList[int]()
	ll.Reverse() // should not panic on empty

	ll.AddAtEnd(1)
	ll.AddAtEnd(2)
	ll.AddAtEnd(3)

	ll.Reverse()

	if ll.Head.Val != 3 || ll.Head.Next.Val != 2 || ll.Head.Next.Next.Val != 1 {
		t.Errorf("Reverse failed")
	}
}

func TestLinkedList_ReversePartition(t *testing.T) {
	ll := NewLinkedList[int]()
	for i := 1; i <= 5; i++ {
		ll.AddAtEnd(i)
	}

	err := ll.ReversePartition(2, 4)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := []int{1, 4, 3, 2, 5}
	cur := ll.Head
	for _, val := range expected {
		if cur == nil || cur.Val != val {
			t.Errorf("Expected %v, got %v", val, cur)
		}
		cur = cur.Next
	}
}

func TestLinkedList_ReversePartition_Errors(t *testing.T) {
	ll := NewLinkedList[int]()
	for i := 1; i <= 3; i++ {
		ll.AddAtEnd(i)
	}

	err := ll.ReversePartition(4, 2)
	if err == nil || err.Error() != "left boundary must smaller than right" {
		t.Errorf("Expected boundary error, got %v", err)
	}

	err = ll.ReversePartition(0, 2)
	if err == nil || err.Error() != "left boundary starts from the first node" {
		t.Errorf("Expected boundary error, got %v", err)
	}

	err = ll.ReversePartition(1, 5)
	if err == nil || err.Error() != "right boundary cannot be greater than the length of the linked list" {
		t.Errorf("Expected boundary error, got %v", err)
	}
}

func TestCheckRangeFromIndex(t *testing.T) {
	ll := NewLinkedList[int]()
	ll.AddAtEnd(1)
	ll.AddAtEnd(2)

	if err := ll.CheckRangeFromIndex(2, 1); err == nil {
		t.Error("Expected error for left > right")
	}
	if err := ll.CheckRangeFromIndex(0, 2); err == nil {
		t.Error("Expected error for left < 1")
	}
	if err := ll.CheckRangeFromIndex(1, 5); err == nil {
		t.Error("Expected error for right > length")
	}
	if err := ll.CheckRangeFromIndex(1, 2); err != nil {
		t.Errorf("Expected no error for valid range, got %v", err)
	}
}

func TestLinkedList_Display(t *testing.T) {
	ll := NewLinkedList[int]()
	ll.AddAtEnd(1)
	ll.AddAtEnd(2)
	ll.AddAtEnd(3)

	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stdout = w

	ll.Display()

	if err := w.Close(); err != nil {
		t.Fatalf("failed to close stdout writer: %v", err)
	}
	os.Stdout = old
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("failed to read stdout: %v", err)
	}

	output := strings.TrimSpace(buf.String())
	expected := "1 2 3"

	if output != expected {
		t.Errorf("Expected output %q, got %q", expected, output)
	}
}
