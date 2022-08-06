package lib

import (
	"fmt"
	"strings"
)

func New() []string {
	return make([]string, 0)
}

func Enqueue(queue []string, element string) []string {
	queue = append(queue, element) // Simply append to enqueue.
	fmt.Println("Enqueued:", element)
	return queue
}

func Dequeue(queue []string) ([]string, string, string) {
	element := queue[0] // The first element is the one to be dequeued.
	info := strings.Split(element, "|")
	fmt.Println("Dequeued:", element)
	return queue[1:], info[0], info[1] // Slice off the element once it is dequeued.
}
