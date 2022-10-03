package main

import "os"

func main() {
	println("here is it")
	f := func(code int) { os.Exit(code) } // want `os.Exit called in main func in main package`
	f(1)
}
