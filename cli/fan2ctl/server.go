package main

import (
	_ "embed"
	"log"
	"net/http"

	"github.com/u3mur4/xiaomi-fan-2/fan2"
)

//go:embed controll.html
var controllHTML []byte

type HandleCmd struct {
	tryFan func() (*fan2.Fan, error)
}

func (handle *HandleCmd) handleCmd(w http.ResponseWriter, r *http.Request) {
	cmd := r.URL.Query().Get("cmd")
	if cmd == "" {
		w.Header().Set("Content-Type", "text/html")
		w.Write(controllHTML)
		return
	}

	fan, err := handle.tryFan()
	if err != nil {
		log.Printf("fan error: %s", err)
		w.Header().Set("Content-Type", "text/html")
		w.Write(controllHTML)
		return
	}
	if !fan.Online() {
		log.Printf("fan is offline")
		w.Header().Set("Content-Type", "text/html")
		w.Write(controllHTML)
		return
	}

	switch cmd {
	case "toggle":
		fan.Toggle()
	case "speed_1":
		fan.SetLevel(fan2.FanLevel1)
	case "speed_2":
		fan.SetLevel(fan2.FanLevel2)
	case "speed_3":
		fan.SetLevel(fan2.FanLevel3)
	case "off_10":
		fan.DelayOff(10)
	case "off_20":
		fan.DelayOff(20)
	case "off_30":
		fan.DelayOff(30)
	case "off_40":
		fan.DelayOff(40)
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write(controllHTML)
}
