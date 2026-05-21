package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/spf13/cobra"
	"github.com/u3mur4/xiaomi-fan-2/fan2"
)

type apiHandler struct {
	origin string
}

func (h *apiHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.origin != "" {
		w.Header().Set("Access-Control-Allow-Origin", h.origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}

	switch {
	case r.Method == "POST" && r.URL.Path == "/api/command":
		h.handleCommand(w, r)
	case r.Method == "GET" && r.URL.Path == "/api/status":
		h.handleStatus(w, r)
	default:
		http.NotFound(w, r)
	}
}

type commandRequest struct {
	Cmd    string `json:"cmd"`
	Device string `json:"device,omitempty"`
}

func (h *apiHandler) resolveFan(name string) (*fan2.Fan, error) {
	return tryFanForDevice(name)
}

func (h *apiHandler) handleCommand(w http.ResponseWriter, r *http.Request) {
	var req commandRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}

	fan, err := h.resolveFan(req.Device)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}
	defer fan.Close()

	var cmdErr error
	switch req.Cmd {
	case "on":
		cmdErr = fan.On()
	case "off":
		cmdErr = fan.Off()
	case "toggle":
		cmdErr = fan.Toggle()
	case "speed_1":
		cmdErr = fan.SetLevel(fan2.FanLevel1)
	case "speed_2":
		cmdErr = fan.SetLevel(fan2.FanLevel2)
	case "speed_3":
		cmdErr = fan.SetLevel(fan2.FanLevel3)
	case "swing":
		var swing bool
		swing, cmdErr = fan.GetHorizontalSwing()
		if cmdErr == nil {
			cmdErr = fan.SetHorizontalSwing(!swing)
		}
	case "mode":
		var mode fan2.Mode
		mode, cmdErr = fan.GetMode()
		if cmdErr == nil {
			cmdErr = fan.SetMode(mode.Toggle())
		}
	case "angle":
		var angle fan2.HorizontalAngle
		angle, cmdErr = fan.GetHorizontalAngle()
		if cmdErr == nil {
			cmdErr = fan.SetHorizontalAngle(angle.Next())
		}
	case "off_10":
		cmdErr = fan.DelayOff(10)
	case "off_20":
		cmdErr = fan.DelayOff(20)
	case "off_30":
		cmdErr = fan.DelayOff(30)
	case "off_40":
		cmdErr = fan.DelayOff(40)
	case "status":
		h.writeStatus(w, fan)
		return
	default:
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "unknown command: " + req.Cmd})
		return
	}

	if cmdErr != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": cmdErr.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *apiHandler) handleStatus(w http.ResponseWriter, r *http.Request) {
	device := r.URL.Query().Get("device")
	fan, err := h.resolveFan(device)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}
	defer fan.Close()
	h.writeStatus(w, fan)
}

func (h *apiHandler) writeStatus(w http.ResponseWriter, fan *fan2.Fan) {
	status, err := fan.GetAllStatus()
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"power":     status.Power,
		"level":     int(status.Level),
		"swing":     status.Swing,
		"angle":     int(status.Angle),
		"mode":      int(status.Mode),
		"led":       status.Brightness,
		"alarm":     status.Alarm,
		"speed":     status.SpeedLevel,
		"childlock": status.ChildLock,
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func init() {
	var port int

	var apiCmd = &cobra.Command{
		Use:     "api",
		Short:   "start REST API server",
		GroupID: "integration",
		Run: func(cmd *cobra.Command, args []string) {
			origin, _ := cmd.Flags().GetString("origin")
			handler := &apiHandler{origin: origin}
			mux := http.NewServeMux()
			mux.Handle("/api/", handler)

			log.Printf("API server listening on :%d", port)
			fmt.Printf("http://0.0.0.0:%d\n", port)
			if err := http.ListenAndServe(fmt.Sprintf(":%d", port), mux); err != nil {
				return
			}
		},
	}
	apiCmd.Flags().IntVar(&port, "port", 35352, "port to listen on")
	apiCmd.Flags().String("origin", "", "allow CORS origin")
	rootCmd.AddCommand(apiCmd)
}
