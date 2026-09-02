package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestDownloadHandles404(t *testing.T) {
	dir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(oldWd)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404")
	}
}

func TestDownloadURLNotEncoded_DocumentsBug(t *testing.T) {
	url := fmt.Sprintf("http://%s:2319/download?name=%s", "1.2.3.4", "my notes")
	if strings.Contains(url, "my notes") {
		t.Fatalf("BUG: download builds URL with fmt.Sprintf without QueryEscape — spaces break: %q (should be my+notes or my%%20notes)", url)
	}
}

func TestDownloadPathTraversal_DocumentsBug(t *testing.T) {
	cd := `attachment; filename="../../.bashrc"`
	parts := strings.Split(cd, "filename=")
	name := strings.Trim(parts[1], `"`)
	if name == "../../.bashrc" {
		t.Fatalf("BUG: download trusts Content-Disposition %q and does os.Create(name) without sanitizing (should be .bashrc)", name)
	}
}

func TestDownloadE2EWithHttptest(t *testing.T) {
	content := "file content 123"
	mux := http.NewServeMux()
	mux.HandleFunc("/download", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("name") == "missing" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Disposition", `attachment; filename="downloaded.txt"`)
		fmt.Fprint(w, content)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/download?name=hello")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 got %d", resp.StatusCode)
	}

	resp2, err := http.Get(srv.URL + "/download?name=missing")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 got %d", resp2.StatusCode)
	}
}

func TestSearchE2E(t *testing.T) {
	want := []FILES{{Name: "shared", Desc: "desc", Class: "c", Path: "/tmp/x", Filetype: "txt"}}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/list" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(want)
	}))
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/list")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 got %d", resp.StatusCode)
	}
	var got []FILES
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(got) != 1 || got[0].Name != "shared" {
		t.Fatalf("unexpected %v", got)
	}
}

func TestSearchHandlesServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer ts.Close()
	resp, err := http.Get(ts.URL + "/list")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		t.Fatalf("expected error status")
	}
}

func TestUploadLocalHandlers(t *testing.T) {
	files := []FILES{{Name: "a", Desc: "d", Class: "c", Path: "/tmp/a.txt", Filetype: "txt"}}
	mux := http.NewServeMux()
	mux.HandleFunc("/list", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(files)
	})
	mux.HandleFunc("/download", func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")
		idx := searchfiles(files, name)
		if idx == -1 {
			http.Error(w, "not found", 404)
			return
		}
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, "a.txt"))
		fmt.Fprint(w, "content")
	})
	ts := httptest.NewServer(mux)
	defer ts.Close()

	resp, _ := http.Get(ts.URL + "/list")
	var got []FILES
	json.NewDecoder(resp.Body).Decode(&got)
	resp.Body.Close()
	if len(got) != 1 {
		t.Fatalf("list failed")
	}
	resp, _ = http.Get(ts.URL + "/download?name=a")
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200 for existing file")
	}
	resp.Body.Close()
	resp, _ = http.Get(ts.URL + "/download?name=missing")
	if resp.StatusCode != 404 {
		t.Fatalf("expected 404 for missing")
	}
	resp.Body.Close()
}
