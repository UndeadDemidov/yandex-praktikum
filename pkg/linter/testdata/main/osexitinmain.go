package main

import (
	"os"
)

func main() {
	println("here is it")
	os.Exit(1) // want `os.Exit called in main func in main package`
}
