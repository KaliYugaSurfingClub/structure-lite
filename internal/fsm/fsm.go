package fsm

import "structure-lite/internal/pages"

type item[T any] struct {
	page  *pages.Page[T]
	space int
}

type FSM[T any] struct {
	items []*item[T]
}

func New[T any]() *FSM[T] {
	return &FSM[T]{
		items: make([]*item[T], 0),
	}
}

func (h *FSM[T]) Push(page *pages.Page[T], space int) {
	i := &item[T]{
		page:  page,
		space: space,
	}

	h.items = append(h.items, i)
	h.up(len(h.items) - 1)
}

func (h *FSM[T]) Pop() *pages.Page[T] {
	if len(h.items) == 0 {
		return nil
	}

	root := h.items[0]

	if root.space == 0 {
		panic("critical: root free space is zero")
	}

	root.space--

	if root.space == 0 {
		h.items[0] = h.items[len(h.items)-1]
		h.items = h.items[:len(h.items)-1]
		h.down(0)
		return root.page
	}

	return root.page
}

func (h *FSM[T]) up(index int) {
	for index > 0 {
		parentIndex := (index - 1) / 2

		if h.items[index].space >= h.items[parentIndex].space {
			break
		}

		h.items[index], h.items[parentIndex] = h.items[parentIndex], h.items[index]
		index = parentIndex
	}
}

func (h *FSM[T]) down(index int) {
	n := len(h.items)

	for {
		leftChild := 2*index + 1
		rightChild := 2*index + 2
		smallest := index

		if leftChild < n && h.items[leftChild].space < h.items[smallest].space {
			smallest = leftChild
		}

		if rightChild < n && h.items[rightChild].space < h.items[smallest].space {
			smallest = rightChild
		}

		if smallest == index {
			break
		}

		h.items[index], h.items[smallest] = h.items[smallest], h.items[index]
		index = smallest
	}
}
