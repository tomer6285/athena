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
	for index, entry := range data {
		if entry.Name == value {
			return index
		}
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
	name, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading name:", err)
		return
	}
	name = strings.TrimSpace(name)

	fmt.Printf("Description: ")
	desc, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading description:", err)
		return
	}
	desc = strings.TrimSpace(desc)

	fmt.Printf("Class: ")
	class, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading class:", err)
		return
	}
	class = strings.TrimSpace(class)

	entry := FILES{
		Name:  name,
		Desc:  desc,
		Class: class,
		Path:  path,
	}

	files = append(files, entry)
	dataBytes, err := json.Marshal(files)
	if err != nil {
		fmt.Println("Error encoding files:", err)
		return
	}
	if err := os.WriteFile(filename, dataBytes, 0644); err != nil {
		fmt.Println("Error saving files:", err)
		return
	}
}

func remove(files []FILES, filename string, value string) {
	index := searchfiles(files, value)
	if index == -1 {
		fmt.Println("Error: Entry does not exist")
		os.Exit(0)
	}
	files = append(files[:index], files[index+1:]...)
	dataBytes, err := json.Marshal(files)
	if err != nil {
		fmt.Println("Error encoding files:", err)
		return
	}
	if err := os.WriteFile(filename, dataBytes, 0644); err != nil {
		fmt.Println("Error saving files:", err)
		return
	}
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
	dataBytes, err := json.Marshal(files)
	if err != nil {
		fmt.Println("Error encoding files:", err)
		return
	}
	if err := os.WriteFile(filename, dataBytes, 0644); err != nil {
		fmt.Println("Error saving files:", err)
		return
	}
	if err := os.Remove(path); err != nil {
		fmt.Println("Error deleting file:", err)
		return
	}
	fmt.Println("File Removed and Deleted Succesfully")
}
