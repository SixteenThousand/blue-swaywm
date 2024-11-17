package main

import (
	"fmt"
	"os"
)

type SwayCtlError struct {
	msg string
}

func (err *SwayCtlError) Error() string {
	return err.msg
}

func NewSwayCtlError(msg string) *SwayCtlError {
	return &SwayCtlError{msg: msg}
}

func MsgAndExit(msg string, err error) {
	fmt.Printf(
		"\033[1;31mError: %s\033[0m\n\nOriginal Error:\n%s\n",
		msg,
		err.Error(),
	)
	os.Exit(1)
}
