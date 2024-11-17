package main

import (
	"encoding/json"
	"fmt"
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

func getCurrentWs() (Workspace, error) {
	var result Workspace
	ipcMsg := exec.Command("swaymsg", "-r", "-t", "get_workspaces")
	ipcStdout, err := ipcMsg.StdoutPipe()
	if err != nil {
		return result, err
	}
	err = ipcMsg.Start()
	if err != nil {
		return result, err
	}
	var workspaces []Workspace
	if err := json.NewDecoder(ipcStdout).Decode(&workspaces); err != nil {
		return result, err
	}
	ipcMsg.Wait()
	for _, ws := range workspaces {
		if ws.Focused {
			return ws, nil
		}
	}
	return result, NewSwayCtlError("No workspace is focused")
}

func getCurrentContainer() Container {
	var current Container
	return current
}

func main() {
	ws, err := getCurrentWs()
	if err != nil {
		MsgAndExit("Sway IPC connection failure", err)
	}
	fmt.Println(ws)
}
