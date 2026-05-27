package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/u3mur4/xiaomi-fan-2/fan2"
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

			var fan *fan2.Fan
			backoff := time.Second
			for {
				if fan == nil {
					fan, _ = tryFan()
				}

				online, power, level, mode := false, false, 0, 0
				if fan != nil {
					p, err := fan.GetPower()
					if err == nil {
						l, _ := fan.GetLevel()
						m, _ := fan.GetMode()
						online, power, level, mode = true, p, int(l), int(m)
					} else {
						fan.Close()
						fan = nil
					}
				}

				printWaybar(online, power, level, mode)

				wait := 30 * time.Second
				if !online {
					wait = backoff
				}

				select {
				case <-sn:
				case <-c:
				case <-time.After(wait):
					if fan != nil {
						fan.Close()
						fan = nil
					}
				}

				if online {
					backoff = time.Second
				} else {
					backoff *= 2
					if backoff > 30*time.Second {
						backoff = 30 * time.Second
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

			var fan *fan2.Fan
			backoff := time.Second
			for {
				if fan == nil {
					fan, _ = tryFan()
				}

				online, power, level := false, false, 0
				if fan != nil {
					p, err := fan.GetPower()
					if err == nil {
						l, _ := fan.GetLevel()
						online, power, level = true, p, int(l)
					} else {
						fan.Close()
						fan = nil
					}
				}

				printPolybar(online, power, level)

				wait := 30 * time.Second
				if !online {
					wait = backoff
				}

				select {
				case <-c:
				case <-time.After(wait):
					if fan != nil {
						fan.Close()
						fan = nil
					}
				}

				if online {
					backoff = time.Second
				} else {
					backoff *= 2
					if backoff > 30*time.Second {
						backoff = 30 * time.Second
					}
				}
			}
		},
	}
	rootCmd.AddCommand(polybarCmd)
}
