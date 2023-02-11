// +build !js

package glitch

// Note: From - "github.com/faiface/mainthread"

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

func mainthreadRun(run func()) {
	callQueue = make(chan func(), CallQueueCap)
	blockingQueue = make(chan func())
	blockingQueueDone = make(chan struct{})

	done := make(chan struct{})
	go func() {
		run()
		done <- struct{}{}
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
		}
	}
}

func mainthreadCall(f func()) {
	checkRun()
	blockingQueue <- f
	<-blockingQueueDone

	// checkRun()
	// done := make(chan struct{})
	// callQueue <- func() {
	// 	f()
	// 	done <- struct{}{}
	// }
	// <-done
}

func mainthreadCallNonBlock(f func()) {
	checkRun()
	callQueue <- f
}

func mainthreadCallErr(f func() error) error {
	checkRun()
	errChan := make(chan error)
	callQueue <- func() {
		errChan <- f()
	}
	return <-errChan
}

/*
import (
	"github.com/faiface/mainthread"
)

func mainthreadRun(run func()) {
	mainthread.Run(run)
}

func mainthreadCall(f func()) {
	mainthread.Call(f)
}

func mainthreadCallNonBlock(f func()) {
	mainthread.CallNonBlock(f)
}

func mainthreadCallErr(f func() error) error {
	return mainthread.CallErr(f)
}
*/

