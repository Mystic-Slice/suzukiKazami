package main

// implement simple queue data structure with a slice
type Queue struct {
	elements []int
}

// Push an element to the end of the queue
func (q *Queue) Push(element int) {
	q.elements = append(q.elements, element)
}

// Push an element to the end of the queue if not already present
func (q *Queue) PushUnique(element int) {
	for _, e := range q.elements {
		if e == element {
			return
		}
	}
	q.elements = append(q.elements, element)
}

// Pop an element from the front of the queue
func (q *Queue) Pop() int {
	if len(q.elements) == 0 {
		return -1
	}
	element := q.elements[0]
	q.elements = q.elements[1:]
	return element
}

// Check if queue is empty
func (q *Queue) IsEmpty() bool {
	return len(q.elements) == 0
}
