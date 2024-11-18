package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"slices"
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

func getCurrentLayout() (Workspace, error) {
	var result Workspace
	result, err := getCurrentWs()
	if err != nil {
		return result, err
	}
	ipcMsg := exec.Command("swaymsg", "-r", "-t", "get_tree")
	ipcStdout, err := ipcMsg.StdoutPipe()
	if err != nil {
		return result, err
	}
	err = ipcMsg.Start()
	if err != nil {
		return result, err
	}
	var data struct {
		Nodes []struct { // outputs
			Type  string
			Name  string // looking for eDP-1
			Nodes []Workspace
		}
	}
	if err := json.NewDecoder(ipcStdout).Decode(&data); err != nil {
		return result, err
	}
	ipcMsg.Wait()
	currentWs, err := getCurrentWs()
	if err != nil {
		return result, err
	}
	for _, wlOutput := range data.Nodes {
		if wlOutput.Name == "eDP-1" {
			for _, ws := range wlOutput.Nodes {
				if ws.Num == currentWs.Num {
					result.Nodes = ws.Nodes
					return result, nil
				}
			}
		}
	}
	return result, NewSwayCtlError("No output (monitor)")
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

func getWindows(layout []Container) []Container {
	var result []Container
	for _, con := range layout {
		if len(con.Nodes) == 0 {
			result = append(result, con)
		} else {
			result = slices.Concat(result, getWindows(con.Nodes))
		}
	}
	return result
}

func getCurrentWindow(layout []Container) (Container, error) {
	for _, con := range layout {
		if con.Focused {
			return con, nil
		}
	}
	return Container{}, NewSwayCtlError("No focused window")
}

func WindowCmd(cmd string, windows []Container, currentId int) int {
	var currentIndex int
	for index, window := range windows {
		if window.Id == currentId {
			currentIndex = index
		}
	}
	numWindows := len(windows)
	var resultIndex int
	switch cmd {
	default:
		return currentId
	case "prev":
		resultIndex = (currentIndex - 1 + numWindows) % numWindows
	case "next":
		resultIndex = (currentIndex + 1 + numWindows) % numWindows
	}
	return windows[resultIndex].Id
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
	case "layout":
		layout, err := getCurrentLayout()
		if err != nil {
			MsgAndExit("Sway IPC connection failure", err)
		}
		fmt.Println(layout)
	case "window":
		layout, err := getCurrentLayout()
		if err != nil {
			MsgAndExit("Sway IPC connection failure", err)
		}
		windows := getWindows(layout.Nodes)
		currentWindow, err := getCurrentWindow(windows)
		if err != nil {
			MsgAndExit("Could not find focused window", err)
		}
		fmt.Println(WindowCmd(os.Args[2], windows, currentWindow.Id))
	case "--version", "-v":
		printVersion()
	}
}
