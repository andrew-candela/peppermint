package internal

import (
	"fmt"
	"runtime"
)

func CheckErrFatal(e error) {
	if e != nil {
		fmt.Println(e)
		panic(e)
	}
}

func LogWithFileName(message string) {
	_, file, line, _ := runtime.Caller(1)
	fmt.Printf("[%s:%d] %s\n", file, line, message)
}
