package main

import (
	"os"
)

func main() {
	println("here is it")
	Exit(1)
}

func Exit(code int) {
	os.Exit(code)
}
