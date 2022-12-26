package recovery

import (
	"fmt"
	"runtime/debug"
)

func Recovery() {
	if r := recover(); r != nil {
		fmt.Println("recovered from ", r)
		debug.PrintStack()
	}
}
