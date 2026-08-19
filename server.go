package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"
)

func uploadlocal(files []FILES) {
	mux := http.NewServeMux()

	mux.HandleFunc("/list", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(files); err != nil {
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

func startServeo() (string, error) {
	cmd := exec.Command("ssh",
		"-o", "StrictHostKeyChecking=no",
		"-R", "80:localhost:2319",
		"serveo.net",
	)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}

	if err := cmd.Start(); err != nil {
		return "", err
	}

	// Read output to find the public URL
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Println(line)

		// Serveo usually prints something like:
		// "Forwarding HTTP traffic from https://xxxx.serveo.net"
		if strings.Contains(line, "https://") {
			parts := strings.Fields(line)
			for _, p := range parts {
				if strings.HasPrefix(p, "https://") {
					return p, nil
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return "", fmt.Errorf("could not get public URL")
}

func uploadout(files []FILES) {
	url, err := startServeo()
	if err != nil {
		fmt.Println("Serveo error:", err)
		return
	}

	fmt.Println("Public URL:", url)
	fmt.Println("Share this with others")
}

func download(ip, filename string) {
	url := fmt.Sprintf("http://%s:2319/download?name=%s", ip, filename)

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()

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
	list(remoteFiles)
}
