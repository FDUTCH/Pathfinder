package pathfind

import (
	"errors"
)

type BinaryHeap struct {
	heap map[int]*Node
	size int
}

func NewBinaryHeap() *BinaryHeap {
	return &BinaryHeap{
		heap: make(map[int]*Node),
		size: 0,
	}
}

func (b *BinaryHeap) Insert(node *Node) (*Node, error) {
	if node.OpenSet() {
		return nil, errors.New("node is already in the heap")
	}

	b.heap[b.size] = node
	node.heapIdx = b.size
	b.upHeap(b.size)
	b.size++
	return node, nil
}

func (b *BinaryHeap) Clear() {
	b.size = 0
}

func (b *BinaryHeap) Peek() *Node {
	if b.size == 0 {
		return nil
	}
	return b.heap[0]
}

func (b *BinaryHeap) Pop() *Node {
	if b.size == 0 {
		return nil
	}

	topNode := b.heap[0]
	b.size--
	b.heap[0] = b.heap[b.size]
	delete(b.heap, b.size)
	if b.size > 0 {
		b.downHeap(0)
	}
	topNode.heapIdx = -1

	return topNode
}

func (b *BinaryHeap) Remove(node *Node) {
	b.size--
	lastNode := b.heap[b.size]
	b.heap[node.heapIdx] = lastNode
	delete(b.heap, b.size)

	if node.heapIdx < b.size {
		if lastNode.f < node.f {
			b.upHeap(node.heapIdx)
		} else {
			b.downHeap(node.heapIdx)
		}
	}
	node.heapIdx = -1
}

func (b *BinaryHeap) ChangeCost(node *Node, newCost float64) {
	oldCost := node.f
	node.f = newCost
	if newCost < oldCost {
		b.upHeap(node.heapIdx)
	} else {
		b.downHeap(node.heapIdx)
	}
}

func (b *BinaryHeap) Size() int {
	return b.size
}

func (b *BinaryHeap) upHeap(index int) {
	node := b.heap[index]
	for index > 0 {
		parentIdx := (index - 1) >> 1
		parent := b.heap[parentIdx]
		if !(node.f < parent.f) {
			break
		}

		b.heap[index] = parent
		parent.heapIdx = index
		index = parentIdx
	}
	b.heap[index] = node
	node.heapIdx = index
}

func (b *BinaryHeap) downHeap(index int) {
	node := b.heap[index]
	currentCost := node.f

	for {
		left := (index << 1) + 1
		right := left + 1
		if left >= b.size {
			break
		}

		minChild := left
		if right < b.size && b.heap[right].f < b.heap[left].f {
			minChild = right
		}

		if b.heap[minChild].f >= currentCost {
			break
		}

		b.heap[index] = b.heap[minChild]
		b.heap[index].heapIdx = index
		index = minChild
	}

	b.heap[index] = node
	node.heapIdx = index
}

func (b *BinaryHeap) IsEmpty() bool {
	return b.size == 0
}

func (b *BinaryHeap) GetHeap() map[int]*Node {
	return b.heap
}
