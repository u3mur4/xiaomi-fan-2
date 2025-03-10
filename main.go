package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path"
	"strconv"
	"syscall"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/buger/jsonparser"
	"github.com/spf13/cobra"
	"github.com/u3mur4/xiaomi-fan-1c/fan1c"
)

type fanConfig struct {
	Location string `json:"location"`
	Id       uint32 `json:"id"`
	Token    string `json:"token"`
}

func getConfigFile() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		log.Fatal(err)
	}

	appDir := path.Join(configDir, "fan1c")
	os.MkdirAll(appDir, 0755)

	return path.Join(appDir, "config.json")
}

func saveConfig(name string, defaultConfig *fanConfig) error {
	location, id, token := "", "", ""
	if defaultConfig != nil {
		location = defaultConfig.Location
		token = defaultConfig.Token
		id = fmt.Sprint(defaultConfig.Id)
	}

	var qs = []*survey.Question{
		{
			Name:     "location",
			Prompt:   &survey.Input{Message: "What is your device location?", Default: location},
			Validate: survey.Required,
		},
		{
			Name:   "id",
			Prompt: &survey.Input{Message: "What is your device id?", Default: id},
		},
		{
			Name:   "token",
			Prompt: &survey.Input{Message: "What is your device token?", Default: token},
		},
	}

	// the answers will be written to this struct
	answers := fanConfig{}

	// perform the questions
	err := survey.Ask(qs, &answers)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(getConfigFile(), os.O_CREATE|os.O_RDONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatal(err)
	}

	if len(data) == 0 {
		data = []byte("{}")
	}

	data, err = jsonparser.Set(data, []byte("\""+answers.Location+"\""), name, "location")
	if err != nil {
		return err
	}

	data, err = jsonparser.Set(data, []byte("\""+answers.Token+"\""), name, "token")
	if err != nil {
		return err
	}

	data, err = jsonparser.Set(data, []byte(fmt.Sprint(answers.Id)), name, "id")
	if err != nil {
		return err
	}

	f2, err := os.Create(getConfigFile())
	if err != nil {
		return err
	}
	defer f2.Close()

	f2.Write(data)
	return nil
}

func getConfig(name string) (*fanConfig, error) {
	f, err := os.OpenFile(getConfigFile(), os.O_CREATE|os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	location, err := jsonparser.GetString(data, name, "location")
	if err != nil {
		return nil, err
	}

	deviceID, err := jsonparser.GetInt(data, name, "id")
	if err != nil {
		return nil, err
	}

	token, err := jsonparser.GetString(data, name, "token")
	if err != nil {
		return nil, err
	}

	return &fanConfig{
		Location: location,
		Id:       uint32(deviceID),
		Token:    token,
	}, nil
}

func exitIfErr(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}
}

func getFan(name string) *fan1c.Fan {
	config, err := getConfig(name)
	exitIfErr(err)

	fan, err := fan1c.NewFan1C(config.Location, config.Id, config.Token)
	fan.Timeout(time.Second)
	// fan.Debug(os.Stderr)
	exitIfErr(err)
	return fan
}

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	var name string
	var rootCmd = &cobra.Command{
		Use:   "fan1c-polybar",
		Short: "control your smart fan",
		Long:  `control your smart fan`,
	}
	rootCmd.PersistentFlags().StringVar(&name, "name", "", "load device from name")

	var editConfig string
	var configCmd = &cobra.Command{
		Use:   "config",
		Short: "list fan config",
		Run: func(cmd *cobra.Command, args []string) {
			if editConfig != "" {
				defaultConfig, _ := getConfig(editConfig)
				err := saveConfig(editConfig, defaultConfig)
				exitIfErr(err)
			} else {
				data, _ := ioutil.ReadFile(getConfigFile())
				var out bytes.Buffer
				json.Indent(&out, data, "", " ")
				fmt.Println(out.String())
				fmt.Println(getConfigFile())
			}
		},
	}
	configCmd.Flags().StringVar(&editConfig, "edit", "", "edit config")
	rootCmd.AddCommand(configCmd)

	var onCmd = &cobra.Command{
		Use:   "on",
		Short: "turn on fan",
		Run: func(cmd *cobra.Command, args []string) {
			err := getFan(name).On()
			exitIfErr(err)
		},
	}
	rootCmd.AddCommand(onCmd)

	var offCmd = &cobra.Command{
		Use:   "off",
		Short: "turn off fan",
		Run: func(cmd *cobra.Command, args []string) {
			err := getFan(name).Off()
			exitIfErr(err)
		},
	}
	rootCmd.AddCommand(offCmd)

	var toogleCmd = &cobra.Command{
		Use:   "toogle",
		Short: "toogle fan",
		Run: func(cmd *cobra.Command, args []string) {
			err := getFan(name).Toogle()
			exitIfErr(err)
		},
	}
	rootCmd.AddCommand(toogleCmd)

	var upCmd = &cobra.Command{
		Use:   "up",
		Short: "increase fan speed",
		Run: func(cmd *cobra.Command, args []string) {
			fan := getFan(name)
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
			fan := getFan(name)
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
			fan := getFan(name)
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
			fan := getFan(name)
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
			fan := getFan(name)
			power, err := fan.GetPower()
			exitIfErr(err)
			level, err := fan.GetLevel()
			exitIfErr(err)
			swing, err := fan.GetHorizontalSwing()
			exitIfErr(err)
			fmt.Printf("Power: %+v\n", power)
			fmt.Printf("Level: %+v\n", level)
			fmt.Printf("Swing: %+v\n", swing)
		},
	}
	rootCmd.AddCommand(statusCmd)

	var swingCmd = &cobra.Command{
		Use:   "swing",
		Short: "toogle horizontal swing",
		Run: func(cmd *cobra.Command, args []string) {
			fan := getFan(name)
			swing, err := fan.GetHorizontalSwing()
			exitIfErr(err)
			err = fan.SetHorizontalSwing(!swing)
			exitIfErr(err)
		},
	}
	rootCmd.AddCommand(swingCmd)

	var polybarCmd = &cobra.Command{
		Use:   "polybar",
		Short: "polybar live output with controll",
		Run: func(cmd *cobra.Command, args []string) {
			c := make(chan os.Signal, 1)
			signal.Notify(c, syscall.SIGUSR1)

			for {
				fan := getFan(name)
				power, _ := fan.GetPower()
				level, _ := fan.GetLevel()
				printPolybar(power, int(level))
				select {
				case <-c:
				case <-time.Tick(time.Second * 30):
				}
			}
		},
	}
	rootCmd.AddCommand(polybarCmd)

	// rootCmd.Flags().BoolVar(&toogle, "toogle", false, "toogle fan power")
	// rootCmd.Flags().BoolVar(&toogle, "update", false, "update server status")
	// rootCmd.Flags().BoolVar(&levelUp, "level-up", false, "increase fan speed")
	// rootCmd.Flags().BoolVar(&levelDown, "level-down", false, "decrease fan speed")
	// rootCmd.Flags().BoolVar(&horizontalSwing, "horizontal-swing", false, "toogle horizontal swing")

	// var saveCmd = &cobra.Command{
	// 	Use:   "save device token",
	// 	Args:  cobra.ExactArgs(2),
	// 	Short: "save token for a device",
	// 	Long:  `save token for a device`,
	// 	Run: func(cmd *cobra.Command, args []string) {
	// 		saveToken(args[0], args[1])
	// 	},
	// }

	// rootCmd.AddCommand(saveCmd)

	// var device string
	// var location string
	// var debug bool

	// 		token, err := getToken(device)
	// 		if err != nil {
	// 			log.Fatal(err)
	// 		}

	// 		deviceNumber, err := strconv.Atoi(device)
	// 		if err != nil {
	// 			log.Fatal(err)
	// 		}

	// 		fan, err := fan1c.NewFan1C(location, uint32(deviceNumber), token)
	// 		if err != nil {
	// 			log.Fatal(err)
	// 		}
	// printPolybar(fanPower, fanLevel)
	// go func() {
	// 	ticker := time.NewTicker(time.Second * 30)

	// 	for {
	// 		select {
	// 		case <-fanPrintNow:
	// 			printPolybar(fanPower, fanLevel)
	// 		case <-ticker.C:
	// 			updateFanProperties(true)
	// 			printPolybar(fanPower, fanLevel)
	// 		}
	// 	}
	// }()

	// }

	// serverCmd.Flags().StringVar(&device, "device", "", "set device id")
	// serverCmd.MarkFlagRequired("device")
	// serverCmd.Flags().StringVar(&location, "location", "", "ip address of the device")
	// serverCmd.MarkFlagRequired("location")
	// serverCmd.Flags().BoolVar(&debug, "debug", false, "enable debug messages")
	// rootCmd.AddCommand(serverCmd)

	rootCmd.Execute()
}
