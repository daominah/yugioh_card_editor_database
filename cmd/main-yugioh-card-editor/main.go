package main

import (
	"log"
	"net/http"

	"github.com/daominah/yugioh_card_editor/pkg/driver/httpsvr"
)

func main() {
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
