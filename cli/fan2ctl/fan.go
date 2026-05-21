package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/u3mur4/xiaomi-fan-2/fan2"
)

func parseOnOff(s string) (bool, error) {
	switch strings.ToLower(s) {
	case "on":
		return true, nil
	case "off":
		return false, nil
	default:
		return false, fmt.Errorf("expected 'on' or 'off', got %q", s)
	}
}

func init() {
	var onCmd = &cobra.Command{
		Use:     "on",
		Short:   "turn on fan",
		GroupID: "fan",
		Run: func(cmd *cobra.Command, args []string) {
			exitIfErr(getFan().On())
		},
	}
	rootCmd.AddCommand(onCmd)

	var offCmd = &cobra.Command{
		Use:     "off",
		Short:   "turn off fan",
		GroupID: "fan",
		Run: func(cmd *cobra.Command, args []string) {
			exitIfErr(getFan().Off())
		},
	}
	rootCmd.AddCommand(offCmd)

	var toggleCmd = &cobra.Command{
		Use:     "toggle",
		Short:   "toggle fan",
		GroupID: "fan",
		Run: func(cmd *cobra.Command, args []string) {
			exitIfErr(getFan().Toggle())
		},
	}
	rootCmd.AddCommand(toggleCmd)

	var upCmd = &cobra.Command{
		Use:     "up",
		Short:   "increase fan speed",
		GroupID: "fan",
		Run: func(cmd *cobra.Command, args []string) {
			fan := getFan()
			level, err := fan.GetLevel()
			exitIfErr(err)
			fan.SetLevel(level.Increase())
		},
	}
	rootCmd.AddCommand(upCmd)

	var downCmd = &cobra.Command{
		Use:     "down",
		Short:   "decrease fan speed",
		GroupID: "fan",
		Run: func(cmd *cobra.Command, args []string) {
			fan := getFan()
			level, err := fan.GetLevel()
			exitIfErr(err)
			fan.SetLevel(level.Decrease())
		},
	}
	rootCmd.AddCommand(downCmd)

	var nextCmd = &cobra.Command{
		Use:     "next",
		Short:   "advance to the next fan speed",
		GroupID: "fan",
		Run: func(cmd *cobra.Command, args []string) {
			fan := getFan()
			level, err := fan.GetLevel()
			exitIfErr(err)
			fan.SetLevel(level.Next())
		},
	}
	rootCmd.AddCommand(nextCmd)

	var delayOffCmd = &cobra.Command{
		Use:     "delay_off <minutes>",
		Args:    cobra.ExactArgs(1),
		Short:   "turn off after the specified minutes",
		GroupID: "fan",
		Run: func(cmd *cobra.Command, args []string) {
			fan := getFan()
			minutes, err := strconv.Atoi(args[0])
			exitIfErr(err)
			exitIfErr(fan.DelayOff(int64(minutes)))
		},
	}
	rootCmd.AddCommand(delayOffCmd)

	var onlineCmd = &cobra.Command{
		Use:     "online",
		Short:   "check if the fan is reachable",
		GroupID: "fan",
		Run: func(cmd *cobra.Command, args []string) {
			getFan().GetPower()
			fmt.Println("online")
		},
	}
	rootCmd.AddCommand(onlineCmd)

	var infoCmd = &cobra.Command{
		Use:     "info",
		Short:   "show device information",
		GroupID: "fan",
		Run: func(cmd *cobra.Command, args []string) {
			info, err := getFan().GetDeviceInformation()
			exitIfErr(err)
			fmt.Printf("Model:        %s\n", info.Model)
			fmt.Printf("Firmware:     %s\n", info.FirmwareVer)
			fmt.Printf("HW Version:   %s\n", info.HardwareVer)
			fmt.Printf("MCU FW:       %s\n", info.MCUFirmwareVer)
			fmt.Printf("MAC:          %s\n", info.MAC)
			fmt.Printf("Uptime:       %ds\n", info.Life)
		},
	}
	rootCmd.AddCommand(infoCmd)

	var statusCmd = &cobra.Command{
		Use:     "status",
		Short:   "show fan status",
		GroupID: "fan",
		Run: func(cmd *cobra.Command, args []string) {
			s, err := getFan().GetAllStatus()
			exitIfErr(err)
			fmt.Printf("Power:     %+v\n", s.Power)
			fmt.Printf("Level:     %+v\n", s.Level)
			fmt.Printf("Swing:     %+v\n", s.Swing)
			fmt.Printf("Angle:     %+v\n", s.Angle)
			fmt.Printf("Mode:      %+v\n", s.Mode)
			fmt.Printf("LED:       %+v\n", s.Brightness)
			fmt.Printf("Alarm:     %+v\n", s.Alarm)
			fmt.Printf("Speed:     %+v\n", s.SpeedLevel)
			fmt.Printf("Childlock: %+v\n", s.ChildLock)
		},
	}
	rootCmd.AddCommand(statusCmd)

	var swingCmd = &cobra.Command{
		Use:     "swing",
		Short:   "toggle horizontal swing",
		GroupID: "fan",
		Run: func(cmd *cobra.Command, args []string) {
			fan := getFan()
			swing, err := fan.GetHorizontalSwing()
			exitIfErr(err)
			exitIfErr(fan.SetHorizontalSwing(!swing))
		},
	}
	rootCmd.AddCommand(swingCmd)

	var modeCmd = &cobra.Command{
		Use:     "mode",
		Short:   "toggle between direct and natural breeze",
		GroupID: "fan",
		Run: func(cmd *cobra.Command, args []string) {
			fan := getFan()
			mode, err := fan.GetMode()
			exitIfErr(err)
			exitIfErr(fan.SetMode(mode.Toggle()))
		},
	}
	rootCmd.AddCommand(modeCmd)

	var angleCmd = &cobra.Command{
		Use:     "angle [<angle>]",
		Short:   "set or cycle horizontal swing angle",
		GroupID: "fan",
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

	var ledCmd = &cobra.Command{
		Use:     "led <on|off>",
		Args:    cobra.ExactArgs(1),
		Short:   "toggle the LED light",
		GroupID: "fan",
		Run: func(cmd *cobra.Command, args []string) {
			status, err := parseOnOff(args[0])
			exitIfErr(err)
			exitIfErr(getFan().SetBrightness(status))
		},
	}
	rootCmd.AddCommand(ledCmd)

	var alarmCmd = &cobra.Command{
		Use:     "alarm <on|off>",
		Args:    cobra.ExactArgs(1),
		Short:   "toggle the beep sound",
		GroupID: "fan",
		Run: func(cmd *cobra.Command, args []string) {
			status, err := parseOnOff(args[0])
			exitIfErr(err)
			exitIfErr(getFan().SetAlarm(status))
		},
	}
	rootCmd.AddCommand(alarmCmd)

	var childlockCmd = &cobra.Command{
		Use:     "childlock <on|off>",
		Args:    cobra.ExactArgs(1),
		Short:   "set physical control lock",
		GroupID: "fan",
		Run: func(cmd *cobra.Command, args []string) {
			status, err := parseOnOff(args[0])
			exitIfErr(err)
			exitIfErr(getFan().SetChildLock(status))
		},
	}
	rootCmd.AddCommand(childlockCmd)

	var speedCmd = &cobra.Command{
		Use:     "speed <1-100>",
		Args:    cobra.ExactArgs(1),
		Short:   "set fine-grained speed level",
		GroupID: "fan",
		Run: func(cmd *cobra.Command, args []string) {
			val, err := strconv.Atoi(args[0])
			exitIfErr(err)
			exitIfErr(getFan().SetSpeedLevel(uint8(val)))
		},
	}
	rootCmd.AddCommand(speedCmd)

	var motorCmd = &cobra.Command{
		Use:     "motor <left|right|no>",
		Args:    cobra.ExactArgs(1),
		Short:   "direct motor movement (momentary)",
		GroupID: "fan",
		Run: func(cmd *cobra.Command, args []string) {
			var val uint8
			switch strings.ToLower(args[0]) {
			case "no":
				val = 0
			case "left":
				val = 1
			case "right":
				val = 2
			default:
				exitIfErr(fmt.Errorf("expected 'left', 'right', or 'no', got %q", args[0]))
			}
			exitIfErr(getFan().SetMotorControl(val))
		},
	}
	rootCmd.AddCommand(motorCmd)
}
