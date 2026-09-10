package stdlib

// demonstration of LinkedList linked list in golang
import (
	"errors"
	"fmt"
)

// Node Structure representing the linkedlist node.
// This node is shared across different implementations.
type Node[T any] struct {
	Val  T
	Prev *Node[T]
	Next *Node[T]
}

// Create new node.
func NewLinkedListNode[T any](val T) *Node[T] {
	return &Node[T]{val, nil, nil}
}

// LinkedList structure with length of the list and its head
type LinkedList[T any] struct {
	length int

	// Note that Node here holds both Next and Prev Node
	// however only the Next node is used in LinkedList methods.
	Head *Node[T]
}

// NewLinkedList returns a new instance of a linked list
func NewLinkedList[T any]() *LinkedList[T] {
	return &LinkedList[T]{}
}

// AddAtBeg adds a new snode with given value at the beginning of the list.
func (ll *LinkedList[T]) AddAtBeg(val T) {
	n := NewLinkedListNode(val)

	n.Next = ll.Head
	ll.Head = n

	ll.length++
}

// AddAtEnd adds a new snode with given value at the end of the list.
func (ll *LinkedList[T]) AddAtEnd(val T) {
	n := NewLinkedListNode(val)

	if ll.Head == nil {
		ll.Head = n
		ll.length++
		return
	}

	cur := ll.Head
	for ; cur.Next != nil; cur = cur.Next {
	}

	cur.Next = n
	ll.length++
}

// DelAtBeg deletes the snode at the head(beginning) of the list
// and returns its value. Returns false if the list is empty.
func (ll *LinkedList[T]) DelAtBeg() (T, bool) {
	if ll.Head == nil {
		var r T
		return r, false
	}

	cur := ll.Head
	ll.Head = cur.Next
	ll.length--

	return cur.Val, true
}

// DelAtEnd deletes the snode at the tail(end) of the list
// and returns its value. Returns false if the list is empty.
func (ll *LinkedList[T]) DelAtEnd() (T, bool) {
	if ll.Head == nil {
		var r T
		return r, false
	}

	if ll.Head.Next == nil {
		return ll.DelAtBeg()
	}

	cur := ll.Head

	for ; cur.Next.Next != nil; cur = cur.Next {
	}

	retval := cur.Next.Val
	cur.Next = nil
	ll.length--
	return retval, true

}

// DelByPos deletes the node at the middle based on position in the list
// and returns its value. Returns false if the list is empty or length is not more than given position
func (ll *LinkedList[T]) DelByPos(pos int) (T, bool) {
	switch {
	case ll.Head == nil:
		var r T
		return r, false
	case pos-1 > ll.length:
		var r T
		return r, false
	case pos-1 == 0:
		return ll.DelAtBeg()
	case pos-1 == ll.Count():
		return ll.DelAtEnd()
	}

	var prev *Node[T]
	var val T
	cur := ll.Head
	count := 0

	for count < pos-1 {
		prev = cur
		cur = cur.Next
		count++
	}

	val = cur.Val
	prev.Next = cur.Next
	ll.length--

	return val, true
}

// Count returns the current size of the list.
func (ll *LinkedList[T]) Count() int {
	return ll.length
}

// Reverse reverses the list.
func (ll *LinkedList[T]) Reverse() {
	var prev, Next *Node[T]
	cur := ll.Head

	for cur != nil {
		Next = cur.Next
		cur.Next = prev
		prev = cur
		cur = Next
	}

	ll.Head = prev
}

// ReversePartition Reverse the linked list from the ath to the bth node
func (ll *LinkedList[T]) ReversePartition(left, right int) error {
	err := ll.CheckRangeFromIndex(left, right)

	if err != nil {
		return err
	}

	tmpNode := &Node[T]{}
	tmpNode.Next = ll.Head
	pre := tmpNode

	for i := 0; i < left-1; i++ {
		pre = pre.Next
	}

	cur := pre.Next
	for i := 0; i < right-left; i++ {
		next := cur.Next
		cur.Next = next.Next
		next.Next = pre.Next
		pre.Next = next
	}

	ll.Head = tmpNode.Next
	return nil
}

func (ll *LinkedList[T]) CheckRangeFromIndex(left, right int) error {
	if left > right {
		return errors.New("left boundary must smaller than right")
	} else if left < 1 {
		return errors.New("left boundary starts from the first node")
	} else if right > ll.length {
		return errors.New("right boundary cannot be greater than the length of the linked list")
	}

	return nil
}

// Display prints out the elements of the list.
func (ll *LinkedList[T]) Display() {
	for cur := ll.Head; cur != nil; cur = cur.Next {
		fmt.Print(cur.Val, " ")
	}

	fmt.Println("")
}
