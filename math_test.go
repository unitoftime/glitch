package glitch

import (
	"fmt"
	"testing"
)

func TestDotProduct(t *testing.T) {
	a := Vec3{1, 2, 3}
	b := Vec3{1, 5, 7}
	dot := a.Dot(b)
	if dot != 32 {
		panic("Should be 32")
	}
}

func TestAngle(t *testing.T) {
	a := Vec3{2, -4, -1}
	b := Vec3{1, 5, 2}
	angle := a.Angle(b)
	// Should be: 2.4928086
	fmt.Println(angle)
}

func TestRotate(t *testing.T) {
	a := Vec3{1, 0, 0}
	b := a.Rotate2D(3.14159/2)
	fmt.Println(b)
}
