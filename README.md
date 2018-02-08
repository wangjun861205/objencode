# objencode
encode object to bytes and decode object from bytes

## usage

import (
  "objencode"
  )
 
 type Test struct {
  Foo string
  Bar int
  }
  
 func main() {
  t := &Test{"hello world", 10}
  b, _ := objencode.Encode(t)
  nt := &Test{}
  _ := objencode.Decode(b, nt)
  fmt.Println(nt.Foo)
  fmt.Println(nt.Bar)
  }
