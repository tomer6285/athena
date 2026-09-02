package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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

func list(files []FILES, filterType string, filter string) {
	// Filter the files first
	var filtered []FILES
	for _, entry := range files {
		if filterType == "-c" && entry.Class != filter {
			continue
		}
		if filterType == "-t" && entry.Filetype != filter {
			continue
		}
		filtered = append(filtered, entry)
	}

	// Find the longest value in each column
	nameWidth := len("Name")
	descWidth := len("Description")
	classWidth := len("Class")
	filetypeWidth := len("Filetype")
	for _, entry := range filtered {
		if len(entry.Name) > nameWidth {
			nameWidth = len(entry.Name)
		}
		if len(entry.Desc) > descWidth {
			descWidth = len(entry.Desc)
		}
		if len(entry.Class) > classWidth {
			classWidth = len(entry.Class)
		}
		if len(entry.Filetype) > filetypeWidth {
			filetypeWidth = len(entry.Filetype)
		}
	}

	// Add padding inside each cell
	nameWidth += 2
	descWidth += 2
	classWidth += 2
	filetypeWidth += 2

	// Helper to center text inside a cell
	center := func(s string, width int) string {
		padding := width - len(s)
		left := padding / 2
		right := padding - left

		return strings.Repeat(" ", left) + s + strings.Repeat(" ", right)
	}

	// Top border
	fmt.Printf("+%s+%s+%s+%s+\n",
		strings.Repeat("-", nameWidth),
		strings.Repeat("-", descWidth),
		strings.Repeat("-", classWidth),
		strings.Repeat("-", filetypeWidth),
	)

	// Header
	fmt.Printf("|%s|%s|%s|%s|\n",
		center("Name", nameWidth),
		center("Description", descWidth),
		center("Class", classWidth),
		center("Filetype", filetypeWidth),
	)

	// Header separator
	fmt.Printf("+%s+%s+%s+%s+\n",
		strings.Repeat("-", nameWidth),
		strings.Repeat("-", descWidth),
		strings.Repeat("-", classWidth),
		strings.Repeat("-", filetypeWidth),
	)

	// File rows
	for _, entry := range filtered {
		fmt.Printf("|%s|%s|%s|%s|\n",
			center(entry.Name, nameWidth),
			center(entry.Desc, descWidth),
			center(entry.Class, classWidth),
			center(entry.Filetype, filetypeWidth),
		)
	}

	// Bottom border
	fmt.Printf("+%s+%s+%s+%s+\n",
		strings.Repeat("-", nameWidth),
		strings.Repeat("-", descWidth),
		strings.Repeat("-", classWidth),
		strings.Repeat("-", filetypeWidth),
	)
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

	filetype := strings.TrimPrefix(filepath.Ext(path), ".")

	entry := FILES{
		Name:     name,
		Desc:     desc,
		Class:    class,
		Path:     path,
		Filetype: filetype,
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

func remove(files []FILES, filename string, toRemove string, delete bool) {
	index := searchfiles(files, toRemove)
	if index == -1 {
		fmt.Println("Error: Entry does not exist")
		os.Exit(0)
	}

	if delete {
		path := files[index].Path
		if err := os.Remove(path); err != nil {
			fmt.Println("Error deleting file:", err)
			return
		}
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

func sync(files []FILES, filename string) {
	var existingFiles []FILES
	for _, entry := range files {
		_, err := os.Stat(entry.Path)
		if !os.IsNotExist(err) {
			existingFiles = append(existingFiles, entry)
		}
	}

	dataBytes, err := json.Marshal(existingFiles)
	if err != nil {
		fmt.Println("Error encoding files:", err)
		return
	}
	if err := os.WriteFile(filename, dataBytes, 0644); err != nil {
		fmt.Println("Error saving files:", err)
		return
	}
}
