package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

type DATA struct {
	Name     string
	Desc     string
	Class    string
	Filetype string
}

func uploadlocal(files []FILES) {

	list := make([]DATA, 0, len(files))
	for _, i := range files {
		list = append(list, DATA{i.Name, i.Desc, i.Class, i.Filetype})
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/list", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(list); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	mux.HandleFunc("/download", func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")
		index := searchfiles(files, name)
		if index == -1 {
			http.Error(w, "file not found", http.StatusNotFound)
			return
		}

		filePath := files[index].Path
		base := path.Base(filePath)
		w.Header().Set("Content-Disposition",
			fmt.Sprintf(`attachment; filename="%s"`, base))
		w.Header().Set("Content-Type", "application/octet-stream")

		http.ServeFile(w, r, filePath)
	})

	server := &http.Server{
		Addr:    ":2319",
		Handler: mux,
	}

	// Wait for Enter in a separate goroutine
	stop := make(chan struct{})

	go func() {
		fmt.Println("Hosting on port 2319...")
		fmt.Println("Press ENTER to stop hosting")

		reader := bufio.NewReader(os.Stdin)
		_, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			fmt.Println("Error reading input:", err)
		}
		close(stop)
	}()

	// Run server (this blocks)
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Println("Server error:", err)
		}
	}()

	// Wait until user presses Enter
	<-stop

	// Shut down server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		fmt.Println("Error shutting down server:", err)
	}
	fmt.Println("Server stopped")
}

func download(ip, filename string) {
	url := fmt.Sprintf("http://%s:2319/download?name=%s", ip, filename)

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		fmt.Println("Error: File does not exist")
		return
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Error: Server returned", resp.Status)
		return
	}

	cd := resp.Header.Get("Content-Disposition")
	var name string
	if cd != "" {
		// Example: attachment; filename="file.pdf"
		parts := strings.Split(cd, "filename=")
		if len(parts) > 1 {
			name = strings.Trim(parts[1], `"`)
		}
	}

	// Fallback: use the requested name
	if name == "" {
		name = filename
	}

	out, err := os.Create(name)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		fmt.Println("Error saving file:", err)
	}
}

func search(ip string) {
	url := fmt.Sprintf("http://%s:2319/list", ip)

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error contacting host:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Remote server error:", resp.Status)
		return
	}

	var remoteFiles []FILES
	err = json.NewDecoder(resp.Body).Decode(&remoteFiles)
	if err != nil {
		fmt.Println("Error decoding response:", err)
		return
	}

	// Reuse your existing printer
	fmt.Println("Files from", ip)
	list(remoteFiles, "", "")
}
