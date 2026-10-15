// Package mainthread runs functions on the main OS thread, which OpenGL and GLFW require.
// The game loop runs directly on the main thread (See: Run), so calls made from it run immediately with no overhead.
// Calls from other goroutines are queued, and run on the main thread at the next Drain (which glitch does once per frame).
package mainthread

import (
	"runtime"
	"sync"
	"sync/atomic"
)

func init() {
	// Note: Locking the OS thread in init pins the main goroutine to the main OS thread
	runtime.LockOSThread()
}

// Reports whether the caller is on the main thread. Until it is set, every caller is treated as being on the main thread
var mainThreadCheck atomic.Pointer[func() bool]

// Sets the function that reports whether the caller is on the main thread (eg once a GL context exists to check against)
func SetMainThreadCheck(check func() bool) {
	mainThreadCheck.Store(&check)
}

func onMainThread() bool {
	check := mainThreadCheck.Load()
	return check == nil || (*check)()
}

// Work queued by other goroutines, which runs at the next Drain
var (
	queueMu  sync.Mutex
	queue    []func()
	draining []func() // Note: Only used by Drain, on the main thread
)

func enqueue(f func()) {
	queueMu.Lock()
	queue = append(queue, f)
	queueMu.Unlock()
}

// Runs the game loop on the main thread. Must be called from the main goroutine
func Run(run func()) {
	run()
}

// Runs f on the main thread and returns once it has finished
// Note: If called from another goroutine, f is queued and runs at the next Drain, so the caller waits until then
func Call(f func()) {
	if onMainThread() {
		f()
		return
	}

	done := make(chan struct{})
	enqueue(func() {
		f()
		close(done)
	})
	<-done
}

// Queues f to run on the main thread at the next Drain, and returns immediately
func CallNonBlock(f func()) {
	enqueue(f)
}

// Like Call, but returns the error from f
func CallErr(f func() error) error {
	var err error
	Call(func() {
		err = f()
	})
	return err
}

// Runs all of the work that other goroutines have queued
// Note: Must be called regularly on the main thread, otherwise callers from other goroutines wait forever
func Drain() {
	queueMu.Lock()
	draining, queue = queue, draining[:0]
	queueMu.Unlock()

	for i := range draining {
		draining[i]()
		draining[i] = nil // Release the function
	}
}
