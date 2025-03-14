package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strconv"
)

const MONITOR = "eDP-1"

type Container struct {
	Id      int
	Name    string
	Focused bool
	Nodes   []Container
}

type Workspace struct {
	Name    string
	Num     int
	Focused bool
	Nodes   []Container
}

func check(err error, msg string) {
	if err != nil {
		fmt.Printf(
			"\x1b[1;31mError: %s\x1b[0m\n%s\n",
			msg,
			err.Error(),
		)
		os.Exit(1)
	}
}

func fail(msg string) {
	fmt.Printf(
		"\x1b[1;31mError: %s\x1b[0m\n",
		msg,
	)
	os.Exit(1)
}

func printHelp() {
	fmt.Println(`
sway extensions (swayext)
Use this to add some nifty little features to sway.

Usage:
	swayext workspace [+/-NUM]
Focuses the workspace with number {CURRENT WORKSPACE NUMBER)+/-NUM,
regardless of whether that workspace exists yet.
	swayext window prev|next
Cycles through windows on current workspace in no particular order,
regardless of layout.
	swayext workspace new
Finds the lowest workspace number not yet taken, and goes to it.
	`)
	os.Exit(0)
}

func getWorkspaces() []Workspace {
	ipcCmd := exec.Command("swaymsg", "-r", "-t", "get_workspaces")
	ipcStdout, err := ipcCmd.StdoutPipe()
	check(err, "getFocusedWs: could not connect to sway IPC")
	defer ipcStdout.Close()
	err = ipcCmd.Start()
	check(err, "getFocusedWs: could not connect to sway IPC")
	var workspaces []Workspace
	err = json.NewDecoder(ipcStdout).Decode(&workspaces)
	check(err, "getFocusedWs: could not connect to sway IPC")
	ipcCmd.Wait()
	return workspaces
}

func getLeaves(tree []Container) []Container {
	var result []Container
	for _, con := range tree {
		if len(con.Nodes) == 0 {
			result = append(result, con)
		} else {
			result = slices.Concat(result, getLeaves(con.Nodes))
		}
	}
	return result
}

func getWindows() []Container {
	ipcCmd := exec.Command("swaymsg", "-r", "-t", "get_tree")
	ipcStdout, err := ipcCmd.StdoutPipe()
	check(err, "getWindows: could not connect to sway IPC")
	err = ipcCmd.Start()
	check(err, "getWindows: could not connect to sway IPC")
	var data struct {
		Nodes []struct { // outputs
			Name              string
			Nodes             []Workspace
			Current_workspace string
		}
	}
	err = json.NewDecoder(ipcStdout).Decode(&data)
	check(err, "getWindows: could not connect to sway IPC")
	ipcCmd.Wait()
	for _, wlOutput := range data.Nodes {
		if wlOutput.Name == MONITOR {
			for _, ws := range wlOutput.Nodes {
				if ws.Name == wlOutput.Current_workspace {
					return getLeaves(ws.Nodes)
				}
			}
			break
		}
	}
	fail("getWindows: no workspace??")
	return make([]Container, 1)
}

func gotoWorkspace(num int) {
	if num < 0 {
		num = 0
	}
	ipcCmd := exec.Command(
		"swaymsg",
		"workspace",
		strconv.Itoa(num),
	)
	err := ipcCmd.Run()
	check(err, "Could not get Sway IPC to change workspace")
}

func gotoWindow(windowId int) {
	ipcCmd := exec.Command(
		"swaymsg",
		fmt.Sprintf("[con_id=%d]", windowId),
		"focus",
	)
	err := ipcCmd.Run()
	check(err, "Could not get Sway IPC to change window")
}

func main() {
	if len(os.Args) <= 2 {
		printHelp()
	}
	if len(os.Args) > 3 {
		fail("Too many arguments!")
	}
	switch os.Args[1] {
	case "workspace":
		switch os.Args[2] {
		case "new":
			var wsNums []int
			for _, ws := range getWorkspaces() {
				wsNums = append(wsNums, ws.Num)
			}
			slices.Sort(wsNums)
			for i := 1; i < len(wsNums); i++ {
				if wsNums[i] > i+1 {
					gotoWorkspace(i + 1)
					return
				}
			}
			gotoWorkspace(len(wsNums) + 1)
		default:
			delta, err := strconv.Atoi(os.Args[2])
			check(err, "swayext workspace NUM: please give an actual (whole) number!")
			focusedWsNum := -1
			for _, ws := range getWorkspaces() {
				if ws.Focused {
					focusedWsNum = ws.Num
					break
				}
			}
			if focusedWsNum < 0 {
				fail("No focused workspace??")
			}
			gotoWorkspace(focusedWsNum + delta)
		}
	case "window":
		focusedIndex := 0
		windows := getWindows()
		for index, window := range windows {
			if window.Focused {
				focusedIndex = index
			}
		}
		numWindows := len(windows)
		resultIndex := focusedIndex
		switch os.Args[2] {
		case "prev":
			resultIndex = (focusedIndex - 1 + numWindows) % numWindows
		case "next":
			resultIndex = (focusedIndex + 1 + numWindows) % numWindows
		}
		gotoWindow(windows[resultIndex].Id)
	default:
		printHelp()
	}
}
