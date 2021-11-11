package main

import (
	"math/rand"
	"reflect"
	"testing"
	"time"
)

func InterfaceSlice(slice interface{}) []interface{} {
	s := reflect.ValueOf(slice)
	if s.Kind() != reflect.Slice {
		panic("InterfaceSlice() given a non-slice type")
	}
	// Keep the distinction between nil and empty slice input
	if s.IsNil() {
		return nil
	}
	ret := make([]interface{}, s.Len())
	for i := 0; i < s.Len(); i++ {
		ret[i] = s.Index(i).Interface()
	}
	return ret
}

type TestStruct struct {
	name string
	age  int
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func randTestStruct(lenArray int, lenMap int) map[int][]TestStruct {
	randomStructMap := make(map[int][]TestStruct, lenMap)
	for i := 0; i < lenMap; i++ {
		var testStructs = make([]TestStruct, 0)
		for k := 0; k < lenArray; k++ {
			rand.Seed(time.Now().UnixNano())
			randomString := randSeq(10)
			randomInt := rand.Intn(100)
			testStructs = append(testStructs, TestStruct{name: randomString, age: randomInt})
		}
		randomStructMap[i] = testStructs
	}
	return randomStructMap
}

// func BenchmarkLoopConversion(b *testing.B) {
// 	var testStructMap = randTestStruct(10, 100)
// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		obj := make([]interface{}, len(testStructMap[i%100]))
// 		for k := range testStructMap[i%100] {
// 			obj[k] = testStructMap[i%100][k]
// 		}
// 	}
// }

func newSlice(size int) []float64 {
	return make([]float64, size)
}

func newSliceGeneric[T any](size int) []T {
	return make([]T, size)
}

func conv(s []float64) interface{} {
	return s[0:5]
}

func convGen[T any](s []T) interface{} {
	return s[0:5]
}

func BenchmarkRegularSlice(b *testing.B) {
	slice := newSlice(1e9)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		iSlice := conv(slice)

		switch s := iSlice.(type) {
		case []float64:
			s[0] = s[0] + 1
		default:
			panic("ERROR")
		}
	}
}

func BenchmarkRegularSliceGeneric(b *testing.B) {
	slice := newSliceGeneric[float64](1e9)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		iSlice := convGen[float64](slice)

		switch s := iSlice.(type) {
		case []float64:
			s[0] = s[0] + 1
		default:
			panic("ERROR")
		}
	}
}

// func BenchmarkRegularSliceConv(b *testing.B) {
// 	slice := make([]float64, b.N)
// 	var iSlice interface{}

// 	b.ResetTimer()

// 	for i := 0; i < b.N; i++ {
// 		iSlice = slice

// 		switch s := iSlice.(type) {
// 		case []float64:
// 			s[0] = s[0] + 1
// 		default:
// 			panic("ERROR")
// 		}
// 	}
// }

// func BenchmarkGenericSliceConv(b *testing.B) {
// 	slice := make([]float64, b.N)
// 	var iSlice interface{}

// 	b.ResetTimer()

// 	for i := 0; i < b.N; i++ {
// 		iSlice = slice

// 		switch s := iSlice.(type) {
// 		case []float64:
// 			s[0] = s[0] + 1
// 		default:
// 			panic("ERROR")
// 		}
// 	}
// }
