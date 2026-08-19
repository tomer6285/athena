package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func searchfiles(data []FILES, value string) int {
	index := 0
	for _, entry := range data {
		if entry.Name == value {
			return index
		}
		index = index + 1
	}
	return -1
}

func list(files []FILES) {
	fmt.Printf("All Shared Files\n")
	fmt.Println("")
	fmt.Printf("%-10s %-15s %-10s %15s\n", "Name", "Description", "Class", "Path")
	fmt.Println("----------------------------------------------------------")
	for _, entry := range files {
		fmt.Printf("%-10v   %-15v   %-10v   %-15v\n", entry.Name, entry.Desc, entry.Class, entry.Path)
	}
}

func open(files []FILES, file string) {
	index := searchfiles(files, file)
	if index == -1 {
		fmt.Println("Error, File not in list")
		return
	}

	path := files[index].Path
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin": // macOS
		cmd = exec.Command("open", path)
	case "linux":
		cmd = exec.Command("xdg-open", path)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "'"+path+"'")
	default:
		fmt.Println("Unsupported OS")
		return
	}

	err := cmd.Start()
	if err != nil {
		fmt.Println("Error opening file:", err)
	}
}

func add(files []FILES, filename string, path string) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("Name: ")
	new_name, _ := reader.ReadString('\n')
	new_name = strings.TrimSpace(new_name)

	fmt.Printf("Description: ")
	new_desc, _ := reader.ReadString('\n')
	new_desc = strings.TrimSpace(new_desc)

	fmt.Printf("Class: ")
	new_class, _ := reader.ReadString('\n')
	new_class = strings.TrimSpace(new_class)

	entry := FILES{
		Name:  new_name,
		Desc:  new_desc,
		Class: new_class,
		Path:  path,
	}

	files = append(files, entry)
	dataBytes, _ := json.Marshal(files)
	_ = os.WriteFile(filename, dataBytes, 0644)
}

func remove(files []FILES, filename string, value string) {
	index := searchfiles(files, value)
	if index == -1 {
		fmt.Println("Error: Entry does not exist")
		os.Exit(0)
	}
	files = append(files[:index], files[index+1:]...)
	dataBytes, _ := json.Marshal(files)
	_ = os.WriteFile(filename, dataBytes, 0644)
	fmt.Println("File Removed Succesfully")
}

func delete(files []FILES, filename string, value string) {
	index := searchfiles(files, value)
	if index == -1 {
		fmt.Println("Error: Entry does not exist")
		os.Exit(0)
	}

	path := files[index].Path
	files = append(files[:index], files[index+1:]...)
	dataBytes, _ := json.Marshal(files)
	_ = os.WriteFile(filename, dataBytes, 0644)
	os.Remove(path)
	fmt.Println("File Removed and Deleted Succesfully")

}
