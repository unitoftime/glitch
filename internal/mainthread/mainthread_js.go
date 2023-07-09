// +build js

package mainthread

// Note: This is an optimization for browsers, which only have one thread.

func Run(run func()) {
	run()
}

func Call(f func()) {
	f()
}

func CallNonBlock(f func()) {
	go f()
}

func CallErr(f func() error) error {
	return f()
}
