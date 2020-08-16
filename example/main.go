// +build example

package main

import (
	"syscall/js"
)

func main() {
	p := js.Global().Get("document").Call("createElement", "p")
	p.Set("innerText", "Hello, World!")
	js.Global().Get("document").Get("body").Call("appendChild", p)
}