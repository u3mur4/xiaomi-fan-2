package main

import (
	_ "embed"
	"fmt"
	"log"
	"net/http"

	"github.com/spf13/cobra"
)

//go:embed controll.html
var controllHTML []byte

func init() {
	var port int

	var webCmd = &cobra.Command{
		Use:     "web",
		Short:   "start web UI",
		GroupID: "integration",
		Run: func(cmd *cobra.Command, args []string) {
			origin, _ := cmd.Flags().GetString("origin")
			api := &apiHandler{origin: origin}

			mux := http.NewServeMux()
			mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/" {
					http.NotFound(w, r)
					return
				}
				w.Header().Set("Content-Type", "text/html")
				w.Write(controllHTML)
			})
			mux.Handle("/api/", api)

			log.Printf("web UI listening on :%d", port)
			fmt.Printf("http://0.0.0.0:%d\n", port)
			if err := http.ListenAndServe(fmt.Sprintf(":%d", port), mux); err != nil {
				return
			}
		},
	}
	webCmd.Flags().IntVar(&port, "port", 8080, "port to listen on")
	webCmd.Flags().String("origin", "", "allow CORS origin")
	rootCmd.AddCommand(webCmd)
}
