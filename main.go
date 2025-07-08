package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"
	"strconv"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/u3mur4/xiaomi-fan-2/fan2"
)

type fanConfig struct {
	Location string `json:"location"`
	Id       uint32 `json:"id"`
	Token    string `json:"token"`
}

func exitIfErr(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}
}

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	viper.SetConfigName("config") // name of config file (without extension)
	viper.SetConfigType("yaml")   // REQUIRED if the config file does not have the extension in the name
	configDir, _ := os.UserConfigDir()
	configDir = path.Join(configDir, "xiaomi-fan-2")
	os.MkdirAll(configDir, 0755)
	viper.AddConfigPath(configDir) // call multiple times to add many search paths
	viper.AddConfigPath(".")       // optionally look for config in the working directory

	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			configFile := path.Join(configDir, "config.yaml")
			if _, err := os.Create(configFile); err != nil { // perm 0666
				fmt.Println("cannot cretate config file: ", configFile)
				exitIfErr(err)
			}

			err = viper.ReadInConfig()
			exitIfErr(err)
		} else {
			// Config file was found but another error was produced
			panic(fmt.Errorf("fatal error config file: %w", err))
		}
	}

	var name string
	var config fanConfig

	var rootCmd = &cobra.Command{
		Use:   "xiaomi-fan-2",
		Short: "control your smart fan",
	}
	rootCmd.PersistentFlags().StringVar(&name, "name", "", "load device from name")
	rootCmd.PersistentFlags().StringVar(&config.Location, "location", "", "location of the fan")
	rootCmd.PersistentFlags().Uint32Var(&config.Id, "id", 0, "id of the fan")
	rootCmd.PersistentFlags().StringVar(&config.Token, "token", "", "token of the fan")

	getFan := func() *fan2.Fan {
		locationKey := "location"
		IdKey := "id"
		tokenKey := "token"
		if name != "" {
			locationKey = fmt.Sprintf("%s.location", name)
			IdKey = fmt.Sprintf("%s.id", name)
			tokenKey = fmt.Sprintf("%s.token", name)
		}

		viper.BindPFlag(locationKey, rootCmd.PersistentFlags().Lookup("location"))
		viper.BindPFlag(IdKey, rootCmd.PersistentFlags().Lookup("id"))
		viper.BindPFlag(tokenKey, rootCmd.PersistentFlags().Lookup("token"))

		config.Location = viper.GetString(locationKey)
		config.Id = viper.GetUint32(IdKey)
		config.Token = viper.GetString(tokenKey)

		fan, err := fan2.NewFan2(config.Location, config.Id, config.Token)
		exitIfErr(err)
		return fan
	}

	var saveCmd = &cobra.Command{
		Use:   "save [NAME]",
		Short: "save location, id, token flag as name",
		Run: func(cmd *cobra.Command, args []string) {
			if config.Id == 0 || config.Location == "" || config.Token == "" || name == "" {
				log.Fatal("you have to specify location, token, id and name flags")
			}

			viper.Set(fmt.Sprintf("%s.location", name), config.Location)
			viper.Set(fmt.Sprintf("%s.id", name), config.Id)
			viper.Set(fmt.Sprintf("%s.token", name), config.Token)
			configFile := viper.ConfigFileUsed()
			err := viper.WriteConfigAs(configFile)
			exitIfErr(err)
			fmt.Println("Config file writen to:", configFile)
		},
	}
	rootCmd.AddCommand(saveCmd)

	var onCmd = &cobra.Command{
		Use:   "on",
		Short: "turn on fan",
		Run: func(cmd *cobra.Command, args []string) {
			err := getFan().On()
			exitIfErr(err)
		},
	}
	rootCmd.AddCommand(onCmd)

	var offCmd = &cobra.Command{
		Use:   "off",
		Short: "turn off fan",
		Run: func(cmd *cobra.Command, args []string) {
			err := getFan().Off()
			exitIfErr(err)
		},
	}
	rootCmd.AddCommand(offCmd)

	var toogleCmd = &cobra.Command{
		Use:   "toogle",
		Short: "toogle fan",
		Run: func(cmd *cobra.Command, args []string) {
			err := getFan().Toogle()
			exitIfErr(err)
		},
	}
	rootCmd.AddCommand(toogleCmd)

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
		Use:   "delay_off [MINUTES]",
		Args:  cobra.ExactArgs(1),
		Short: "turn off after the specified minutes",
		Run: func(cmd *cobra.Command, args []string) {
			fan := getFan()
			minutes, err := strconv.ParseInt(args[0], 10, 64)
			exitIfErr(err)
			err = fan.DelayOff(minutes)
			exitIfErr(err)
		},
	}
	rootCmd.AddCommand(delayOffCmd)

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
			fmt.Printf("Power: %+v\n", power)
			fmt.Printf("Level: %+v\n", level)
			fmt.Printf("Swing: %+v\n", swing)
			fmt.Printf("Angle: %+v\n", angle)
		},
	}
	rootCmd.AddCommand(statusCmd)

	var swingCmd = &cobra.Command{
		Use:   "swing",
		Short: "toogle horizontal swing",
		Run: func(cmd *cobra.Command, args []string) {
			fan := getFan()
			swing, err := fan.GetHorizontalSwing()
			exitIfErr(err)
			err = fan.SetHorizontalSwing(!swing)
			exitIfErr(err)
		},
	}
	rootCmd.AddCommand(swingCmd)

	var angleCmd = &cobra.Command{
		Use:   "angle [ANGLE]",
		Short: "next horizontal swing angle",
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
			err = fan.SetHorizontalAngle(angle)
			exitIfErr(err)
		},
	}
	rootCmd.AddCommand(angleCmd)

	var waybarCmd = &cobra.Command{
		Use:   "waybar",
		Short: "waybar live output with controll",
		Run: func(cmd *cobra.Command, args []string) {
			c := make(chan os.Signal, 1)
			signal.Notify(c, syscall.SIGUSR1)

			for {
				fan := getFan()
				power, _ := fan.GetPower()
				level, _ := fan.GetLevel()
				printWaybar(power, int(level))
				select {
				case <-c:
				case <-time.Tick(time.Second * 30):
				}
				fan.Close()
			}
		},
	}
	rootCmd.AddCommand(waybarCmd)

	var polybarCmd = &cobra.Command{
		Use:   "polybar",
		Short: "polybar live output with controll",
		Run: func(cmd *cobra.Command, args []string) {
			c := make(chan os.Signal, 1)
			signal.Notify(c, syscall.Signal(34))
			signal.Notify(c, syscall.Signal(35))

			for {
				fan := getFan()
				power, _ := fan.GetPower()
				level, _ := fan.GetLevel()
				printPolybar(power, int(level))
				select {
				case <-c:
				case <-time.Tick(time.Second * 30):
				}
				fan.Close()
			}
		},
	}
	rootCmd.AddCommand(polybarCmd)

	var serverCmd = &cobra.Command{
		Use:   "server",
		Short: "controll with http server",
		Run: func(cmd *cobra.Command, args []string) {
			port := 35352
			handler := HandleCmd{getFan: getFan}
			http.HandleFunc("/", handler.handleCmd)
			fmt.Printf("http://0.0.0.0:%d\n", port)
			if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
				return
			}
		},
	}
	rootCmd.AddCommand(serverCmd)

	rootCmd.Execute()
}
