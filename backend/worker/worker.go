package worker

import (
	Structs "hueify/structs"
	"sync"
	"time"
)

type Work func(chan Structs.PotentialAlbum)

type WorkerPool struct {
	Jobs      *chan Structs.PotentialAlbum
	WaitGroup sync.WaitGroup
}

func CreateWorkerPool(numWorkers int) *WorkerPool {
	ch := make(chan Structs.PotentialAlbum)
	wp := &WorkerPool{
		Jobs: &ch,
	}
	for i := 0; i < numWorkers; i++ {
		wp.WaitGroup.Add(1)
	}
	return wp
}

func (wp *WorkerPool) Worker(work func()) {
	defer wp.WaitGroup.Done()
	for {
		select {
		case _, ok := <-*wp.Jobs:
			if ok {
				work()
			} else {
				return
			}
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func (wp *WorkerPool) Close() {
	close(*wp.Jobs)
	wp.WaitGroup.Wait()
}
