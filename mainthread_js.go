// +build js
package glitch

// Note: This is an optimization for browsers, which only have one thread.

func mainthreadRun(run func()) {
	run()
}

func mainthreadCall(f func()) {
	f()
}

func mainthreadCallNonBlock(f func()) {
	go f()
}

func mainthreadCallErr(f func() error) error {
	return f()
}
