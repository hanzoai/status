package static

import (
	"io/fs"
	"strings"
	"testing"
)

func TestEmbed(t *testing.T) {
	scenarios := []struct {
		path                  string
		shouldExist           bool
		expectedContainString string
	}{
		{
			path:                  "index.html",
			shouldExist:           true,
			expectedContainString: "</body>",
		},
		{
			path:                  "favicon.ico",
			shouldExist:           true,
			expectedContainString: "", // not checking because it's an image
		},
		{
			path:        "file-that-does-not-exist.html",
			shouldExist: false,
		},
	}
	staticFileSystem, err := fs.Sub(FileSystem, RootPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, scenario := range scenarios {
		t.Run(scenario.path, func(t *testing.T) {
			content, err := fs.ReadFile(staticFileSystem, scenario.path)
			if !scenario.shouldExist {
				if err == nil {
					t.Errorf("%s should not have existed", scenario.path)
				}
			} else {
				if err != nil {
					t.Errorf("opening %s should not have returned an error, got %s", scenario.path, err.Error())
				}
				if len(content) == 0 {
					t.Errorf("%s should have existed in the static FileSystem, but was empty", scenario.path)
				}
				if len(scenario.expectedContainString) > 0 && !strings.Contains(string(content), scenario.expectedContainString) {
					t.Errorf("%s should have contained %s, but did not", scenario.path, scenario.expectedContainString)
				}
			}
		})
	}

	// Verify Next.js static assets directory exists
	t.Run("_next/static/chunks", func(t *testing.T) {
		entries, err := fs.ReadDir(staticFileSystem, "_next/static/chunks")
		if err != nil {
			t.Errorf("_next/static/chunks directory should exist, got error: %s", err.Error())
			return
		}
		if len(entries) == 0 {
			t.Error("_next/static/chunks directory should not be empty")
		}
		// Verify at least one JS file exists
		hasJS := false
		for _, entry := range entries {
			if strings.HasSuffix(entry.Name(), ".js") {
				hasJS = true
				break
			}
		}
		if !hasJS {
			t.Error("_next/static/chunks should contain at least one .js file")
		}
	})

	t.Run("_next/static/css", func(t *testing.T) {
		entries, err := fs.ReadDir(staticFileSystem, "_next/static/css")
		if err != nil {
			t.Errorf("_next/static/css directory should exist, got error: %s", err.Error())
			return
		}
		if len(entries) == 0 {
			t.Error("_next/static/css directory should not be empty")
		}
	})
}
