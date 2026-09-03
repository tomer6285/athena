package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type FILES struct {
	Name     string
	Desc     string
	Class    string
	Path     string
	Filetype string
}

func help() {
	fmt.Println("Athena — peer-to-peer file sharing")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  athena <command> [arguments]")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Printf("  %-24s %s\n", "list [-t, -c]", "List all shared files (-c/t lets you filter by class/filetype)")
	fmt.Printf("  %-24s %s\n", "open <name>", "Open a file by name")
	fmt.Printf("  %-24s %s\n", "add <path>", "Add a file to the shared list")
	fmt.Printf("  %-24s %s\n", "remove [-d] <name>", "Remove from list (-d also deletes from disk)")
	fmt.Printf("  %-24s %s\n", "upload-local", "Host files on the local network (port 2319)")
	fmt.Printf("  %-24s %s\n", "upload-out", "Host files publicly via serveo.net tunnel")
	fmt.Printf("  %-24s %s\n", "search <ip>", "Browse files shared by another user")
	fmt.Printf("  %-24s %s\n", "download <ip> <name>", "Download a file from another user")
	fmt.Println("")
	fmt.Println("Options:")
	fmt.Printf("  %-24s %s\n", "-h, --help, help", "Show this help message")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  athena list")
	fmt.Println("  athena add ./notes.pdf")
	fmt.Println("  athena open notes")
	fmt.Println("  athena remove -d notes")
	fmt.Println("  athena upload-local")
	fmt.Println("  athena search 192.168.1.10")
	fmt.Println("  athena download 192.168.1.10 notes")
}

func main() {
	// Check if files exist and creates them if needed
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error getting home directory:", err)
		os.Exit(1)
	}
	path := home + "/.config/athena"
	file := path + "/files.json"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, os.ModePerm); err != nil {
			fmt.Println("Error creating directory:", err)
			os.Exit(1)
		}
	}
	if _, err := os.Stat(file); os.IsNotExist(err) {
		if err := os.WriteFile(file, []byte("[]"), 0644); err != nil {
			fmt.Println("Error creating file:", err)
			os.Exit(1)
		}
	}

	// Reads files into memory
	read, err := os.ReadFile(file)
	if err != nil {
		fmt.Println("Error reading file:", err)
		os.Exit(1)
	}
	var files []FILES
	if err := json.Unmarshal(read, &files); err != nil {
		fmt.Println("Error parsing json:", err)
		os.Exit(1)
	}

	if len(os.Args) < 2 {
		help()
		os.Exit(0)
	}

	switch os.Args[1] {

	case "list":
		if len(os.Args) < 3 {
			list(files, "", "")
			return
		}

		if os.Args[2] != "-c" && os.Args[2] != "-t" {
			fmt.Println("Error, filter type not supported")
			os.Exit(0)
		}

		if len(os.Args) < 4 {
			fmt.Println("Error, no filter inserted")
			os.Exit(0)
		}

		list(files, os.Args[2], strings.ToLower(os.Args[3]))

	case "open":
		if len(os.Args) < 3 {
			fmt.Println("Error, please enter the name of the file you want to open")
			os.Exit(0)
		}

		open(files, os.Args[2])

	case "add":
		if len(os.Args) < 3 {
			fmt.Println("Please provide the path to the file you would like to add")
			os.Exit(0)
		}

		add(files, file, strings.Join(os.Args[2:], " "))

	case "remove":
		if len(os.Args) < 3 || (os.Args[2] == "-d" && len(os.Args) < 4) {
			fmt.Println("Error, please provide a file to remove")
			os.Exit(0)
		}

		delete := false
		toRemove := strings.Join(os.Args[2:], " ")

		if os.Args[2] == "-d" {
			delete = true
			toRemove = strings.Join(os.Args[3:], " ")
		}

		remove(files, file, toRemove, delete)

	case "sync":
		sync(files, file)

	case "upload-local":
		uploadlocal(files)

	case "search":
		if len(os.Args) < 3 {
			fmt.Println("Error, please provide the ip address you want to index")
			os.Exit(0)
		}

		search(os.Args[2])

	case "download":
		if len(os.Args) < 4 {
			fmt.Println("Error, please provide the ip address and file you'd like to download")
			os.Exit(0)
		}

		download(os.Args[2], strings.Join(os.Args[3:], " "))

	case "-h", "--help", "help":
		help()

	default:
		fmt.Println("")
		fmt.Println("ERROR: Invalid Command")
	}
}
