package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
)

type Workspace struct {
	Name    string
	Num     int
	Focused bool
	Nodes   []Container
}

type Container struct {
	Id      int
	Name    string
	Focused bool
	Nodes   []Container
}

type SwayCtlError struct {
	msg           string
	originalError error
}

func (scErr *SwayCtlError) Error() string {
	return fmt.Sprintf(
		"Error: %s\n\nOriginal Error:\n%s",
		scErr.msg,
		scErr.originalError.Error(),
	)
}

func Wrap(err error, msg string) *SwayCtlError {
	return &SwayCtlError{
		msg:           msg,
		originalError: err,
	}
}

func (scErr *SwayCtlError) Exit() {
	fmt.Printf(
		"\033[1;31mError: %s\033[0m\n\nOriginal Error:\n%s",
		scErr.msg,
		scErr.originalError.Error(),
	)
	os.Exit(1)
}

func getCurrentWs() (Workspace, error) {
	var result Workspace
	const ipcErrMsg string = "Could not connect to Sway IPC"
	ipcMsg := exec.Command("swaymsg", "-r", "-t", "get_workspaces")
	output, err := ipcMsg.StdoutPipe()
	if err != nil {
		return result, Wrap(err, ipcErrMsg)
	}
	err = ipcMsg.Run()
	if err != nil {
		return result, Wrap(err, ipcErrMsg)
	}
	var rawOutput []byte
	_, err = output.Read(rawOutput)
	if err != nil {
		return result, Wrap(err, ipcErrMsg)
	}
	var workspaces []Workspace
	err = json.Unmarshal(rawOutput, &workspaces)
	if err != nil {
		return result, Wrap(err, "Could not parse JSON returned by sway IPC")
	}
	for _, ws := range workspaces {
		if ws.Focused {
			return ws, nil
		}
	}
	return result, Wrap(nil, "No focused workspace")
}

func getCurrentContainer() Container {
	var current Container
	return current
}

func ErrorMsg(err error, msg string) {
	if err == nil {
		return
	}
	fmt.Printf("\033[1;31mError:%s\033[0m\n\n", msg)
	fmt.Println(err)
	os.Exit(1)
}

func main() {
	fmt.Println("Hello, World!")
}
