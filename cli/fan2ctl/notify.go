package main

import (
	"fmt"
	"log"
	"net/http"
)

type Notifier struct {
	Port int
}

func (n *Notifier) Notify() {
	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/notify", n.Port))
	if err != nil {
		return
	}
	resp.Body.Close()
}

func (n *Notifier) Listen() (<-chan struct{}, error) {
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

		if err := http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", n.Port), nil); err != nil {
			log.Fatal(err)
		}
	}()

	return ch, nil
}
