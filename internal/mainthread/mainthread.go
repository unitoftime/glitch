// +build !js

package mainthread

// Note: Adapted from - "github.com/faiface/mainthread"
import (
	"errors"
	"runtime"
)

var CallQueueCap = 16

var (
	callQueue chan func()
	blockingQueue chan func()
	blockingQueueDone chan struct{}
)

func init() {
	runtime.LockOSThread()
}

func checkRun() {
	if callQueue == nil {
		panic(errors.New("mainthread: did not call Run"))
	}
}

func Run(run func()) {
	callQueue = make(chan func(), CallQueueCap)
	blockingQueue = make(chan func())
	blockingQueueDone = make(chan struct{})

	// panicQueue := make(chan any) // Used to bubble panics from the run function to the caller
	done := make(chan error) // TODO - maybe return errors?
	go func() {
		// defer func() {
		// 	r := recover()
		// 	if r != nil {
		// 		panicQueue <- r
		// 	}
		// }()

		run()
		done <- nil // TODO - maybe run should return errors
	}()

	for {
		select {
		case f := <-blockingQueue:
			f()
			blockingQueueDone <- struct{}{}
		case f := <-callQueue:
			f()
		case <-done:
			return
		// case p := <-panicQueue:
		// 	panic(p)
		}
	}
}

func Call(f func()) {
	checkRun()
	blockingQueue <- f
	<-blockingQueueDone
}

func CallNonBlock(f func()) {
	checkRun()
	callQueue <- f
}

func CallErr(f func() error) error {
	checkRun()
	errChan := make(chan error)
	callQueue <- func() {
		errChan <- f()
	}
	return <-errChan
}
