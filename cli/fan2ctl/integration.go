package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/u3mur4/xiaomi-fan-2/fan2"
)

type runLoopConfig struct {
	signals     []os.Signal
	useNotifier bool
	output      func(online bool, power bool, level int, mode int)
}

func runLoop(cfg runLoopConfig) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, cfg.signals...)

	var notifCh <-chan struct{}
	if cfg.useNotifier {
		var err error
		notifCh, err = notifier.Listen()
		exitIfErr(err)
	}

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

		cfg.output(online, power, level, mode)

		wait := 30 * time.Second
		if !online {
			wait = backoff
		}

		select {
		case <-notifCh:
		case <-sigCh:
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
}

func init() {
	var waybarCmd = &cobra.Command{
		Use:     "waybar",
		Short:   "waybar live output with control",
		GroupID: "integration",
		Run: func(cmd *cobra.Command, args []string) {
			runLoop(runLoopConfig{
				signals:     []os.Signal{syscall.SIGUSR1},
				useNotifier: true,
				output: func(online bool, power bool, level int, mode int) {
					printWaybar(online, power, level, mode)
				},
			})
		},
	}
	waybarCmd.Flags().IntVar(&notifier.Port, "notify-server-port", notifier.Port, "port to listen for notifications")
	rootCmd.AddCommand(waybarCmd)

	var polybarCmd = &cobra.Command{
		Use:     "polybar",
		Short:   "polybar live output with control",
		GroupID: "integration",
		Run: func(cmd *cobra.Command, args []string) {
			runLoop(runLoopConfig{
				signals: []os.Signal{syscall.Signal(34), syscall.Signal(35)},
				output: func(online bool, power bool, level int, mode int) {
					printPolybar(online, power, level, mode)
				},
			})
		},
	}
	rootCmd.AddCommand(polybarCmd)
}
