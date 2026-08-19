package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type FILES struct {
	Name  string
	Desc  string
	Class string
	Path  string
}

func main() {
	//Check if files exist and creates them if needed
	home, _ := os.UserHomeDir()
	path := (home + "/.config/athena")
	file := (path + "/files.json")
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		os.Mkdir(path, os.ModePerm)
	}
	_, err = os.Stat(file)
	if os.IsNotExist(err) {
		f, _ := os.Create(file)
		f.WriteString("[]")
		f.Close()
	}

	//Reads files into memory
	read, _ := os.ReadFile(file)
	var files []FILES
	_ = json.Unmarshal(read, &files)

	if len(os.Args) < 2 {
		fmt.Println("Athena Commands:")
		fmt.Println("")
		fmt.Println("list - List all files")
		fmt.Println("open - opens a file")
		fmt.Println("add - add a file")
		fmt.Println("remove - remove a file from list but does not delete")
		fmt.Println("delete - remove a file from the list and delete it")
		fmt.Println("upload-local - set to upload mode on the local network")
		fmt.Println("upload-out - set to upload mode outbound to anywhere")
		fmt.Println("search - search another user's list")
		fmt.Println("download - download a file from another user")
		os.Exit(0)
	}

	if os.Args[1] == "list" {
		list(files)
	}

	if os.Args[1] == "open" {
		if len(os.Args) < 3 {
			fmt.Println("Error, please enter the name of the file you want to open")
			os.Exit(0)
		}
		open(files, os.Args[2])
	}

	if os.Args[1] == "add" {
		if len(os.Args) < 3 {
			fmt.Println("Please provide the path to the file you would like to add")
			os.Exit(0)
		}
		add(files, file, os.Args[2])
	}

	if os.Args[1] == "remove" {
		if len(os.Args) < 3 {
			fmt.Println("Error, please provide a file to remove")
			os.Exit(0)
		}
		remove(files, file, os.Args[2])
		read, _ := os.ReadFile(file)
		var files []FILES
		_ = json.Unmarshal(read, &files)
		list(files)
	}

	if os.Args[1] == "upload-local" {
		uploadlocal(files)
	}

	if os.Args[1] == "upload-out" {
		uploadout(files)
	}

	if os.Args[1] == "search" {
		if len(os.Args) < 3 {
			fmt.Println("Error, please provide the ip address you want to index")
			os.Exit(0)
		}
		search(os.Args[2])
	}

	if os.Args[1] == "download" {
		if len(os.Args) < 4 {
			fmt.Println("Error, please provide the ip address and file you'd like to download")
			os.Exit(0)
		}
		download(os.Args[2], os.Args[3])
	}

	//fmt.Println("")
	//fmt.Println("ERROR: Invalid Command")
}
