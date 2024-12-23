package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"slices"
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
	fmt.Println("swayctl, version 0.0.1.scratch")
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

func WriteWorkspacesIcon() (string, error) {
	// find workspace icon location
	var WORKSPACES_ICON_PATH string
	if len(os.Getenv("XDG_CONFIG_HOME")) != 0 {
		WORKSPACES_ICON_PATH = os.Getenv("XDG_CONFIG_HOME") + "/sway/state/workspaces.svg"
	} else {
		WORKSPACES_ICON_PATH = os.Getenv("HOME") + "/.config/sway/state/workspaces.svg"
	}
	// get workspaces
	ipcMsg := exec.Command("swaymsg", "-r", "-t", "get_workspaces")
	ipcStdout, err := ipcMsg.StdoutPipe()
	if err != nil {
		return WORKSPACES_ICON_PATH, err
	}
	err = ipcMsg.Start()
	if err != nil {
		return WORKSPACES_ICON_PATH, err
	}
	var workspaces []Workspace
	if err := json.NewDecoder(ipcStdout).Decode(&workspaces); err != nil {
		return WORKSPACES_ICON_PATH, err
	}
	ipcMsg.Wait()
	// construct SVG
	svg := `<svg
        xmlns="http://www.w3.org/2000/svg"
        viewBox="0 0 90 90"
    >`
	var row, col int
	var useTag string
	const (
		focused_rgba  = "#3cc86499"
		nonempty_rgba = "#3cc86466"
	)
	for _, ws := range workspaces {
		col = (ws.Num - 1) % 3
		row = (ws.Num - 1) / 3
		if ws.Focused {
			useTag = fmt.Sprintf(
				`<rect x="%d" y="%d" fill="%s" height="30" width="30" />`,
				col*30,
				row*30,
				focused_rgba,
			)
		} else {
			useTag = fmt.Sprintf(
				`<rect x="%d" y="%d" fill="%s" height="30" width="30" />`,
				col*30,
				row*30,
				nonempty_rgba,
			)
		}
		svg += useTag
	}
	svg += "</svg>"
	// write file
	fp, err := os.Create(WORKSPACES_ICON_PATH)
	if err != nil {
		return WORKSPACES_ICON_PATH, err
	}
	_, err = fp.Write([]byte(svg))
	if err != nil {
		return WORKSPACES_ICON_PATH, err
	}
	return WORKSPACES_ICON_PATH, nil
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
	case "tile":
		layout, err := getCurrentLayout()
		if err != nil {
			MsgAndExit("Sway IPC connection failure", err)
		}
		windows := getWindows(layout.Nodes)
		switch len(windows) {
		default:
			fmt.Println(":")
		case 2:
			fmt.Println("swaymsg split vertical")
		}
	case "ws-icon":
		path, err := WriteWorkspacesIcon()
		if err != nil {
			MsgAndExit("Could not make icon", err)
		}
		fmt.Println(path)
	case "--version", "-v":
		printVersion()
	}
}
