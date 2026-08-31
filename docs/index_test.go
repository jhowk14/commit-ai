package docs

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestWebsiteUsesCurrentReleaseAndSupportedPlatforms(t *testing.T) {
	mainSource, err := os.ReadFile(filepath.Join("..", "cmd", "commit-ai", "main.go"))
	if err != nil {
		t.Fatal(err)
	}
	match := regexp.MustCompile(`var version = "([^"]+)"`).FindStringSubmatch(string(mainSource))
	if len(match) != 2 {
		t.Fatal("versão do binário não encontrada")
	}

	page, err := os.ReadFile("index.html")
	if err != nil {
		t.Fatal(err)
	}
	content := string(page)
	for _, expected := range []string{
		`<html lang="en">`,
		`<link rel="canonical" href="https://jhowk14.github.io/commit-ai/">`,
		`"@type": "SoftwareApplication"`,
		`"codeRepository": "https://github.com/jhowk14/commit-ai"`,
		`"softwareVersion": "` + match[1] + `"`,
		"Linux and macOS",
		"Windows",
		"AI-powered Git commit messages",
		"v" + match[1] + "/any-linux/install.sh",
		"v" + match[1] + "/windows/install.ps1",
	} {
		if !strings.Contains(content, expected) {
			t.Fatalf("site não contém %q", expected)
		}
	}
}
