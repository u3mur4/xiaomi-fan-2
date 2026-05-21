package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

func init() {
	var waybarCmd = &cobra.Command{
		Use:     "waybar",
		Short:   "waybar live output with control",
		GroupID: "integration",
		Run: func(cmd *cobra.Command, args []string) {
			c := make(chan os.Signal, 1)
			signal.Notify(c, syscall.SIGUSR1)
			sn, err := notifier.Listen()
			exitIfErr(err)

			fan, _ := tryFan()
			for {
				if fan == nil {
					fan, _ = tryFan()
				}
				if fan != nil {
					power, err := fan.GetPower()
					if err != nil {
						printWaybar(false, false, 0, 0)
						fan.Close()
						fan = nil
					} else {
						level, _ := fan.GetLevel()
						mode, _ := fan.GetMode()
						printWaybar(true, power, int(level), int(mode))
					}
				} else {
					printWaybar(false, false, 0, 0)
				}
				select {
				case <-sn:
				case <-c:
				case <-time.Tick(time.Second * 30):
					if fan != nil {
						fan.Close()
						fan = nil
					}
				}
			}
		},
	}
	waybarCmd.Flags().IntVar(&notifier.Port, "notify-server-port", notifier.Port, "port to listen for notifications")
	rootCmd.AddCommand(waybarCmd)

	var polybarCmd = &cobra.Command{
		Use:     "polybar",
		Short:   "polybar live output with control",
		GroupID: "integration",
		Run: func(cmd *cobra.Command, args []string) {
			c := make(chan os.Signal, 1)
			signal.Notify(c, syscall.Signal(34))
			signal.Notify(c, syscall.Signal(35))

			fan, _ := tryFan()
			for {
				if fan == nil {
					fan, _ = tryFan()
				}
				if fan != nil {
					power, err := fan.GetPower()
					if err != nil {
						printPolybar(false, false, 0)
						fan.Close()
						fan = nil
					} else {
						level, _ := fan.GetLevel()
						printPolybar(true, power, int(level))
					}
				} else {
					printPolybar(false, false, 0)
				}
				select {
				case <-c:
				case <-time.Tick(time.Second * 30):
					if fan != nil {
						fan.Close()
						fan = nil
					}
				}
			}
		},
	}
	rootCmd.AddCommand(polybarCmd)
}
