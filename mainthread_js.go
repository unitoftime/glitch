// +build js

package glitch

// TODO - Would this be faster?
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
