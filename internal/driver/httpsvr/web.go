package httpsvr

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"

	"github.com/mywrap/gofast"
)

func NewHandlerGUI(webDirPath string) (http.Handler, error) {
	if webDirPath == "" {
		projectRoot, err := gofast.GetProjectRootGit()
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
