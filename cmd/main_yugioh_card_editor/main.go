package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/daominah/yugioh_card_editor/internal/driver/httpsvr"
)

func main() {
	log.SetFlags(log.Lshortfile)
	log.SetOutput(customLogger{})

	go func() {
		listenPort := ":20808"
		handler, err := httpsvr.NewHandlerGUI("")
		if err != nil {
			log.Fatalf("error NewHandlerGUI: %v", err)
		}
		log.Printf("serving user interface on http://localhost%v", listenPort)
		err = http.ListenAndServe(listenPort, handler)
		if err != nil {
			log.Fatalf("error ListenAndServe: %v", err)
		}
	}()

	select {}
}

// customLogger adds time to the beginning of each log line, write to stdout
type customLogger struct{}

func (writer customLogger) Write(bytes []byte) (int, error) {
	return fmt.Printf("%v %s", time.Now().UTC().Format("2006-01-02T15:04:05.000Z07:00"), bytes)
}
