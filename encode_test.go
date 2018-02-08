package objencode

import (
	"fmt"
	"testing"
)

type Test struct {
	// X bool
	// X int
	// X string
	// Y string
	// Z map[string]string
	X []byte
	Y string
}

func TestEncode(t *testing.T) {
	// test := Test{"test", "hello world", map[string]string{"a": "1", "b": "2"}}
	test := Test{[]byte{1, 2, 3}, "hello world"}
	b, _ := Encode(&test)
	fmt.Println(b)
	newTest := &Test{}
	err := Decode(b, newTest)
	if err != nil {
		panic(err)
	}
	fmt.Println(newTest.X[1])
	fmt.Println(newTest.Y)
	// var a [24]byte
	// copy(a[:], b)
	// ptr := unsafe.Pointer(&a)
	// newTest := *(*Test)(ptr)
	// fmt.Println(newTest.X)
}
