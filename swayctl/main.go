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

func printVersion() {
	fmt.Println("swayctl, version 0.0.0")
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

func WorkspaceCmd(cmd string, current int) int {
	current = current + 8
	switch cmd {
	default:
		return 1
	case "prev":
		current -= 1
	case "next":
		current += 1
	case "down":
		current += 3
	case "up":
		current -= 3
	}
	return (current % 9) + 1
}

func main() {
	switch os.Args[1] {
	default:
		printVersion()
	case "workspace":
		ws, err := getCurrentWs()
		if err != nil {
			MsgAndExit("Sway IPC connection failure", err)
		}
		fmt.Println(WorkspaceCmd(os.Args[2], ws.Num))
	case "--version", "-v":
		printVersion()
	}
}
