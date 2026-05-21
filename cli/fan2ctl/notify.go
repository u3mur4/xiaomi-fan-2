package main

import (
	"fmt"
	"log"
	"net/http"
)

func notifyServer(port int) (<-chan struct{}, error) {
	ch := make(chan struct{})

	go func() {
		http.HandleFunc("/notify", func(w http.ResponseWriter, r *http.Request) {
			select {
			case ch <- struct{}{}:
			default:
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprint(w, "Error sending notification")
			}
		})

		if err := http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", port), nil); err != nil {
			log.Fatal(err)
		}
	}()

	return ch, nil
}
