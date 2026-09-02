package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func captureStdout(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	f()
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	os.Stdout = old
	return buf.String()
}

func sampleFiles() []FILES {
	return []FILES{
		{Name: "notes", Desc: "math notes", Class: "CS101", Path: "/tmp/notes.pdf", Filetype: "pdf"},
		{Name: "photo", Desc: "vacation", Class: "Personal", Path: "/tmp/photo.jpg", Filetype: "jpg"},
		{Name: "slides", Desc: "lecture", Class: "CS101", Path: "/tmp/slides.pdf", Filetype: "pdf"},
	}
}

func TestSearchFilesFound(t *testing.T) {
	files := sampleFiles()
	if idx := searchfiles(files, "notes"); idx != 0 {
		t.Fatalf("expected 0 got %d", idx)
	}
	if idx := searchfiles(files, "photo"); idx != 1 {
		t.Fatalf("expected 1 got %d", idx)
	}
}

func TestSearchFilesNotFound(t *testing.T) {
	if idx := searchfiles(sampleFiles(), "missing"); idx != -1 {
		t.Fatalf("expected -1 got %d", idx)
	}
}

func TestSearchFilesCaseSensitive(t *testing.T) {
	if idx := searchfiles(sampleFiles(), "Notes"); idx != -1 {
		t.Fatalf("expected case-sensitive miss, got %d", idx)
	}
}

func TestSearchFilesEmpty(t *testing.T) {
	if idx := searchfiles(nil, "notes"); idx != -1 {
		t.Fatalf("expected -1 for empty slice")
	}
}

func TestListNoFilter(t *testing.T) {
	out := captureStdout(func() { list(sampleFiles(), "", "") })
	for _, name := range []string{"notes", "photo", "slides"} {
		if !strings.Contains(out, name) {
			t.Fatalf("list output missing %q: %s", name, out)
		}
	}
	if !strings.Contains(out, "Name") || !strings.Contains(out, "Description") {
		t.Fatalf("header missing: %s", out)
	}
}

func TestListFilterByClass(t *testing.T) {
	out := captureStdout(func() { list(sampleFiles(), "-c", "CS101") })
	if !strings.Contains(out, "notes") || !strings.Contains(out, "slides") {
		t.Fatalf("expected CS101 entries: %s", out)
	}
	if strings.Contains(out, "photo") {
		t.Fatalf("should not contain photo: %s", out)
	}
}

func TestListFilterByType(t *testing.T) {
	out := captureStdout(func() { list(sampleFiles(), "-t", "pdf") })
	if !strings.Contains(out, "notes") || !strings.Contains(out, "slides") {
		t.Fatalf("expected pdf entries: %s", out)
	}
	if strings.Contains(out, "photo") {
		t.Fatalf("should not contain jpg: %s", out)
	}
}

func TestListFilterCaseMismatch(t *testing.T) {
	out := captureStdout(func() { list(sampleFiles(), "-c", "cs101") })
	if strings.Contains(out, "notes") {
		t.Fatalf("current code is case-sensitive, lowercased filter should not match CS101: %s", out)
	}
}

func TestListEmpty(t *testing.T) {
	out := captureStdout(func() { list(nil, "", "") })
	if !strings.Contains(out, "Name") {
		t.Fatalf("empty list should still print header: %s", out)
	}
}

func TestSyncRemovesMissing(t *testing.T) {
	dir := t.TempDir()
	existing := filepath.Join(dir, "keep.txt")
	os.WriteFile(existing, []byte("hi"), 0644)
	missing := filepath.Join(dir, "gone.txt")

	files := []FILES{
		{Name: "keep", Desc: "k", Class: "c", Path: existing, Filetype: "txt"},
		{Name: "gone", Desc: "g", Class: "c", Path: missing, Filetype: "txt"},
	}
	outFile := filepath.Join(dir, "files.json")
	sync(files, outFile)
	data, _ := os.ReadFile(outFile)
	var got []FILES
	json.Unmarshal(data, &got)
	if len(got) != 1 || got[0].Name != "keep" {
		t.Fatalf("expected only keep, got %v", got)
	}
}

func TestSyncKeepsAllWhenExist(t *testing.T) {
	dir := t.TempDir()
	p1 := filepath.Join(dir, "a.txt")
	p2 := filepath.Join(dir, "b.txt")
	os.WriteFile(p1, []byte("a"), 0644)
	os.WriteFile(p2, []byte("b"), 0644)
	files := []FILES{
		{Name: "a", Desc: "", Class: "", Path: p1, Filetype: "txt"},
		{Name: "b", Desc: "", Class: "", Path: p2, Filetype: "txt"},
	}
	outFile := filepath.Join(dir, "files.json")
	sync(files, outFile)
	var got []FILES
	data, _ := os.ReadFile(outFile)
	json.Unmarshal(data, &got)
	if len(got) != 2 {
		t.Fatalf("expected 2 kept, got %d", len(got))
	}
}

func TestSyncEmptyInput(t *testing.T) {
	dir := t.TempDir()
	outFile := filepath.Join(dir, "files.json")
	sync(nil, outFile)
	data, _ := os.ReadFile(outFile)
	var got []FILES
	json.Unmarshal(data, &got)
	if len(got) != 0 {
		t.Fatalf("expected empty, got %v", got)
	}
}

func TestRemoveExisting(t *testing.T) {
	dir := t.TempDir()
	outFile := filepath.Join(dir, "files.json")
	files := sampleFiles()
	data, _ := json.Marshal(files)
	os.WriteFile(outFile, data, 0644)
	remove(files, outFile, "photo", false)
	var got []FILES
	raw, _ := os.ReadFile(outFile)
	json.Unmarshal(raw, &got)
	if len(got) != 2 {
		t.Fatalf("expected 2 after remove, got %d", len(got))
	}
	if searchfiles(got, "photo") != -1 {
		t.Fatalf("photo should be removed")
	}
}

func TestRemoveWithDeleteFlag(t *testing.T) {
	dir := t.TempDir()
	tmpFile := filepath.Join(dir, "deleteme.txt")
	os.WriteFile(tmpFile, []byte("bye"), 0644)
	outFile := filepath.Join(dir, "files.json")
	files := []FILES{{Name: "deleteme", Desc: "", Class: "", Path: tmpFile, Filetype: "txt"}}
	data, _ := json.Marshal(files)
	os.WriteFile(outFile, data, 0644)
	remove(files, outFile, "deleteme", true)
	if _, err := os.Stat(tmpFile); !os.IsNotExist(err) {
		t.Fatalf("file should be deleted from disk")
	}
	var got []FILES
	raw, _ := os.ReadFile(outFile)
	json.Unmarshal(raw, &got)
	if len(got) != 0 {
		t.Fatalf("expected empty after delete")
	}
}

func TestAddWritesFile(t *testing.T) {
	dir := t.TempDir()
	outFile := filepath.Join(dir, "files.json")
	os.WriteFile(outFile, []byte("[]"), 0644)
	tmpPath := filepath.Join(dir, "input.pdf")
	os.WriteFile(tmpPath, []byte("pdf"), 0644)

	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	w.WriteString("myname\nmydesc\nmyclass\n")
	w.Close()
	os.Stdin = r
	var files []FILES
	add(files, outFile, tmpPath)
	os.Stdin = oldStdin

	raw, _ := os.ReadFile(outFile)
	var got []FILES
	json.Unmarshal(raw, &got)
	if len(got) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(got))
	}
	if got[0].Name != "myname" || got[0].Filetype != "pdf" || got[0].Path != tmpPath {
		t.Fatalf("unexpected entry: %+v", got[0])
	}
}
