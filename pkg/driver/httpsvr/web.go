package httpsvr

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"

	"github.com/daominah/yugioh_card_editor_database/pkg/base"
)

func NewHandlerGUI(webDirPath string) (http.Handler, error) {
	if webDirPath == "" {
		projectRoot, err := base.GetProjectRootDir()
		if err != nil {
			return nil, fmt.Errorf("empty webDirPath and cannot GetProjectRootDir: %w", err)
		}
		webDirPath = filepath.Join(projectRoot, "web")
		log.Printf("empty path for web app static directory, use the default location: %v", webDirPath)
	}
	handler := http.NewServeMux()
	handler.Handle("/", http.FileServer(http.Dir(webDirPath)))
	return handler, nil
}
