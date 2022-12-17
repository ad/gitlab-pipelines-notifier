package main

import (
	"fmt"
	"runtime/debug"
)

func recovery() {
	if r := recover(); r != nil {
		fmt.Println("recovered from ", r)
		debug.PrintStack()
	}
}
