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

	// Verify Vite static assets directory exists
	t.Run("assets", func(t *testing.T) {
		entries, err := fs.ReadDir(staticFileSystem, "assets")
		if err != nil {
			t.Errorf("assets directory should exist, got error: %s", err.Error())
			return
		}
		if len(entries) == 0 {
			t.Error("assets directory should not be empty")
		}
		hasJS := false
		hasCSS := false
		for _, entry := range entries {
			if strings.HasSuffix(entry.Name(), ".js") {
				hasJS = true
			}
			if strings.HasSuffix(entry.Name(), ".css") {
				hasCSS = true
			}
		}
		if !hasJS {
			t.Error("assets should contain at least one .js file")
		}
		if !hasCSS {
			t.Error("assets should contain at least one .css file")
		}
	})
}
