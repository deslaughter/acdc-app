package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestFiles(t *testing.T) {

	// Find all FAST files
	matches, err := filepath.Glob("testdata/*-ED.fst")
	if err != nil {
		t.Fatal(err)
	}

	for _, filePath := range matches {

		files, err := ParseFiles(filePath)
		if err != nil {
			t.Fatalf("error parsing '%s': %s", filePath, err)
		}

		b, err := json.MarshalIndent(files, "", "\t")
		if err != nil {
			t.Fatal(err)
		}

		// Get directory
		dir := filepath.Base(filepath.Dir(filePath))

		path := filepath.Join("testdata", "output", dir)
		if err := os.MkdirAll(path, 0777); err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(filepath.Join(path, "files.json"), b, 0777); err != nil {
			t.Fatal(err)
		}

		if err := files.Write(path, ""); err != nil {
			t.Fatal(err)
		}
	}
}
