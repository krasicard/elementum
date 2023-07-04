//go:build shared

package main

import (
	"C"

	"bufio"
	"fmt"
	"os"

	"github.com/elgatito/elementum/exit"
)

func initShared() {
	exit.Reset()
	exit.IsShared = true
}

func initLog(arg string) {
	logPath = arg

	originalStdout := os.Stdout

	r, w, _ := os.Pipe()

	os.Stdout = w
	os.Stderr = w

	go func() {
		logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
		if err != nil {
			fmt.Printf("Could not open log file '%s' for writing: %s\n", logPath, err)
			return
		}
		defer logFile.Close()

		os.Stdout.WriteString(fmt.Sprintf("Redirecting Stdout/Stderr to %s\r\n", logPath))
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			s := scanner.Text() + "\r\n"
			logFile.WriteString(s)
			originalStdout.WriteString(s)
		}
	}()
}

//export start
func start() {
	initShared()

	main()
}

//export startWithLog
func startWithLog(log *C.char) int {
	initShared()
	initLog(C.GoString(log))

	main()

	return exit.Code
}

//export startWithArgs
func startWithArgs(args *C.char) int {
	initShared()

	exit.Args = C.GoString(args)
	main()

	return exit.Code
}

//export startWithLogAndArgs
func startWithLogAndArgs(log, args *C.char) int {
	initShared()
	initLog(C.GoString(log))

	exit.Args = C.GoString(args)
	main()

	return exit.Code
}
