package objencode

import (
	"fmt"
	"goginx/http"
	"testing"
)

func TestEncode(t *testing.T) {
	test := &http.HTTPRequest{
		// RemoteAddr: [4]byte{127, 0, 0, 1},
		// RemotePort: 12345,
		Method:  "GET",
		URI:     "/",
		Version: "1.1",
		Headers: make(map[string][]string),
		Body:    []byte{},
		Fd:      123,
	}
	test.Headers["Connection"] = []string{}
	// test.Headers["User-Agent"] = []string{"User-Agent:Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/64.0.3282.119 Safari/537.36"}
	b, _ := Encode(test)
	fmt.Println(b)
	newTest := &http.HTTPRequest{}
	err := Decode(b, newTest)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(newTest.RemotePort)
}
