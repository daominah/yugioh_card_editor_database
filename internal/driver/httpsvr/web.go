package httpsvr

import (
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"
)

func NewHandlerGUI(webDirPath string) (http.Handler, error) {
	if webDirPath == "" {
		projectRoot, err := getProjectRootGit()
		if err != nil {
			return nil, fmt.Errorf("empty webDirPath and cannot getProjectRootGit: %v", err)
		}
		webDirPath = filepath.Join(projectRoot, "web")
		log.Printf("empty path for web app static directory, use the default location: %v", webDirPath)
	}
	handler := http.NewServeMux()
	handler.Handle("/", http.FileServer(http.Dir(webDirPath)))
	return handler, nil
}

// getProjectRootGit returns absolute path of the project root dir (base on git)
func getProjectRootGit() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	stdout, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(stdout)), nil
}
