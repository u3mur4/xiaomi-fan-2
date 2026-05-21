package main

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/u3mur4/xiaomi-fan-2/fan2"
)

func init() {
	var onCmd = &cobra.Command{
		Use:   "on",
		Short: "turn on fan",
		Run: func(cmd *cobra.Command, args []string) {
			exitIfErr(getFan().On())
		},
	}
	rootCmd.AddCommand(onCmd)

	var offCmd = &cobra.Command{
		Use:   "off",
		Short: "turn off fan",
		Run: func(cmd *cobra.Command, args []string) {
			exitIfErr(getFan().Off())
		},
	}
	rootCmd.AddCommand(offCmd)

	var toggleCmd = &cobra.Command{
		Use:   "toggle",
		Short: "toggle fan",
		Run: func(cmd *cobra.Command, args []string) {
			exitIfErr(getFan().Toggle())
		},
	}
	rootCmd.AddCommand(toggleCmd)

	var upCmd = &cobra.Command{
		Use:   "up",
		Short: "increase fan speed",
		Run: func(cmd *cobra.Command, args []string) {
			fan := getFan()
			level, err := fan.GetLevel()
			exitIfErr(err)
			fan.SetLevel(level.Increase())
		},
	}
	rootCmd.AddCommand(upCmd)

	var downCmd = &cobra.Command{
		Use:   "down",
		Short: "decrease fan speed",
		Run: func(cmd *cobra.Command, args []string) {
			fan := getFan()
			level, err := fan.GetLevel()
			exitIfErr(err)
			fan.SetLevel(level.Decrease())
		},
	}
	rootCmd.AddCommand(downCmd)

	var nextCmd = &cobra.Command{
		Use:   "next",
		Short: "advance to the next fan speed",
		Run: func(cmd *cobra.Command, args []string) {
			fan := getFan()
			level, err := fan.GetLevel()
			exitIfErr(err)
			fan.SetLevel(level.Next())
		},
	}
	rootCmd.AddCommand(nextCmd)

	var delayOffCmd = &cobra.Command{
		Use:   "delay_off <minutes>",
		Args:  cobra.ExactArgs(1),
		Short: "turn off after the specified minutes",
		Run: func(cmd *cobra.Command, args []string) {
			fan := getFan()
			minutes, err := strconv.Atoi(args[0])
			exitIfErr(err)
			exitIfErr(fan.DelayOff(int64(minutes)))
		},
	}
	rootCmd.AddCommand(delayOffCmd)

	var onlineCmd = &cobra.Command{
		Use:   "online",
		Short: "check if the fan is reachable",
		Run: func(cmd *cobra.Command, args []string) {
			getFan().GetPower()
			fmt.Println("online")
		},
	}
	rootCmd.AddCommand(onlineCmd)

	var statusCmd = &cobra.Command{
		Use:   "status",
		Short: "show fan status",
		Run: func(cmd *cobra.Command, args []string) {
			fan := getFan()
			power, err := fan.GetPower()
			exitIfErr(err)
			level, err := fan.GetLevel()
			exitIfErr(err)
			swing, err := fan.GetHorizontalSwing()
			exitIfErr(err)
			angle, err := fan.GetHorizontalAngle()
			exitIfErr(err)
			mode, err := fan.GetMode()
			exitIfErr(err)
			fmt.Printf("Power: %+v\n", power)
			fmt.Printf("Level: %+v\n", level)
			fmt.Printf("Swing: %+v\n", swing)
			fmt.Printf("Angle: %+v\n", angle)
			fmt.Printf("Mode: %+v\n", mode)
		},
	}
	rootCmd.AddCommand(statusCmd)

	var swingCmd = &cobra.Command{
		Use:   "swing",
		Short: "toggle horizontal swing",
		Run: func(cmd *cobra.Command, args []string) {
			fan := getFan()
			swing, err := fan.GetHorizontalSwing()
			exitIfErr(err)
			exitIfErr(fan.SetHorizontalSwing(!swing))
		},
	}
	rootCmd.AddCommand(swingCmd)

	var modeCmd = &cobra.Command{
		Use:   "mode",
		Short: "toggle between direct and natural breeze",
		Run: func(cmd *cobra.Command, args []string) {
			fan := getFan()
			mode, err := fan.GetMode()
			exitIfErr(err)
			exitIfErr(fan.SetMode(mode.Toggle()))
		},
	}
	rootCmd.AddCommand(modeCmd)

	var angleCmd = &cobra.Command{
		Use:   "angle [<angle>]",
		Short: "set or cycle horizontal swing angle",
		Run: func(cmd *cobra.Command, args []string) {
			fan := getFan()
			angle, _ := fan.GetHorizontalAngle()
			if len(args) > 0 {
				angleInt, err := strconv.Atoi(args[0])
				exitIfErr(err)
				angle = fan2.HorizontalAngle(angleInt)
			} else {
				angle = angle.Next()
			}
			exitIfErr(fan.SetHorizontalAngle(angle))
		},
	}
	rootCmd.AddCommand(angleCmd)
}
