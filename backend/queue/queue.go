package queue

import (
	"fmt"
	"strconv"
	"strings"
)

func New() [][]string {
	return make([][]string, 0)
}

func Enqueue(queue [][]string, element []string) [][]string {
	queue = append(queue, element) // Simply append to enqueue.
	fmt.Println("Enqueued:", element)
	return queue
}

func Dequeue(queue [][]string) ([][]string, string, string, int) {
	element := queue[0] // The first element is the one to be dequeued.
	info := strings.Split(element[0], "|")
	depth, err := strconv.Atoi(element[1])
	if err != nil {
		return queue, "", "", -1
	}
	fmt.Println("Dequeued:", element)
	return queue[1:], info[0], info[1], depth // Slice off the element once it is dequeued.
}
