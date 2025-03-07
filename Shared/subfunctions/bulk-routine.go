package subfunctions

import (
	"os"
	"strconv"
	"time"
)

type BulkRoutineInterface[T any] interface {
	Insert(object T)
}

type BulkRoutine[T any] struct {
	objects      []T
	insert       chan T
	routine      func([]T, any) error
	routineDelay time.Duration
}

func (b *BulkRoutine[T]) Insert(object T) {
	b.insert <- object
}

type BulkRoutineParams[T any] struct {
	Routine        func([]T, any) error
	TransferParams any //Params that you want to pass to the routine.
}

// basic usage.
// call this function, then call Insert() on the returned object to insert objects.
// the routine will be called with the objects in the queue when the queue is full or the delay is reached.
// if you provide it any transfer params, those will also be passed to the routine.
// for example, you can set T to be a tuple of an entity, and a response handler.
// the response handler can be used to send responses back to the client for each response, while you work on the gathered entities using bulk operations.
func NewBulkRoutine[T any](params BulkRoutineParams[T]) BulkRoutineInterface[T] {
	maxQueueSize, err := strconv.Atoi(os.Getenv("MAX_DB_INSERT_COUNT"))
	if err != nil {
		println("Error getting max insert count: ", err.Error())
		panic(err)
	}
	routineDelay, err := strconv.Atoi(os.Getenv("BULK_ROUTINE_DELAY"))
	if err != nil {
		println("Error getting bulk routine delay: ", err.Error())
		panic(err)
	}
	b := BulkRoutine[T]{
		routine:      params.Routine,
		objects:      make([]T, 0, maxQueueSize),
		insert:       make(chan T, maxQueueSize),
		routineDelay: time.Duration(routineDelay) * time.Millisecond,
	}
	go func(passParams any) {
		for {
			initialRequest := <-b.insert
			b.objects = append(b.objects, initialRequest)
			timer := time.NewTimer(b.routineDelay)
		inner:
			for {
				select {
				case object := <-b.insert:
					b.objects = append(b.objects, object)
				case <-timer.C: //wait duration.
					break inner
				}
			}
			if len(b.objects) > 0 {
				batch := append([]T(nil), b.objects...)
				go func(batchCopy []T, passParams any) {
					if err := b.routine(batchCopy, passParams); err != nil {
						println("Error in bulk routine:", err.Error())
					}
				}(batch, passParams)
				b.objects = b.objects[:0]
			}
		}
	}(params.TransferParams)
	return &b
}
