package main

import (
	_ "embed"
	"net/http"

	"github.com/u3mur4/xiaomi-fan-2/fan2"
)

//go:embed controll.html
var controllHTML []byte

type HandleCmd struct {
	getFan func() *fan2.Fan
}

func (handle *HandleCmd) handleCmd(w http.ResponseWriter, r *http.Request) {
	cmd := r.URL.Query().Get("cmd")
			
	if cmd == "toggle" {
		handle.getFan().Toogle()
	} else if cmd == "swing" {
		// swing, _ := handle.getFan().GetHorizontalSwing()
		// handle.getFan().SetHorizontalSwing(!swing)
	} else if cmd == "speed_1" {
		handle.getFan().SetLevel(fan2.FanLevel1)
	} else if cmd == "speed_2" {
		handle.getFan().SetLevel(fan2.FanLevel2)
	} else if cmd == "speed_3" {
		handle.getFan().SetLevel(fan2.FanLevel3)
	} else if cmd == "off_10" {
		handle.getFan().DelayOff(10)
	} else if cmd == "off_20" {
		handle.getFan().DelayOff(20)
	} else if cmd == "off_30" {
		handle.getFan().DelayOff(30)
	} else if cmd == "off_40" {
		handle.getFan().DelayOff(40)
	}

	// Write the file content as the response
	w.Header().Set("Content-Type", "text/html")
	w.Write(controllHTML)

}
