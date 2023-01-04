// // +build !js

package glitch

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
