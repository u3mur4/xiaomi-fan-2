package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/u3mur4/xiaomi-fan-2/fan2"
)

type fanConfig struct {
	Location string `json:"location"`
	Id       uint32 `json:"id"`
	Token    string `json:"token"`
}

var (
	deviceName   string
	config       fanConfig
	rootCmd      = &cobra.Command{
		Use:   "fan2ctl",
		Short: "control your smart fan",
	}
)

func exitIfErr(err error) {
	if err != nil {
		if strings.Contains(err.Error(), "timeout") {
			fmt.Fprintln(os.Stderr, "fan is offline")
		} else {
			fmt.Fprintln(os.Stderr, err)
		}
		sendNotify(notifServerPort)
		os.Exit(-1)
	}
}

func initConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	configDir, _ := os.UserConfigDir()
	configDir = path.Join(configDir, "xiaomi-fan-2")
	os.MkdirAll(configDir, 0755)
	viper.AddConfigPath(configDir)
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			configFile := path.Join(configDir, "config.yaml")
			if _, err := os.Create(configFile); err != nil {
				fmt.Println("cannot create config file: ", configFile)
				exitIfErr(err)
			}
			err = viper.ReadInConfig()
			exitIfErr(err)
		} else {
			panic(fmt.Errorf("fatal error config file: %w", err))
		}
	}
}

func tryFan() (*fan2.Fan, error) {
	locationKey := "location"
	IdKey := "id"
	tokenKey := "token"

	if deviceName != "" {
		locationKey = fmt.Sprintf("%s.location", deviceName)
		IdKey = fmt.Sprintf("%s.id", deviceName)
		tokenKey = fmt.Sprintf("%s.token", deviceName)
	} else {
		defaultName := viper.GetString("_default")
		if defaultName != "" {
			locationKey = fmt.Sprintf("%s.location", defaultName)
			IdKey = fmt.Sprintf("%s.id", defaultName)
			tokenKey = fmt.Sprintf("%s.token", defaultName)
		}
	}

	viper.BindPFlag(locationKey, rootCmd.PersistentFlags().Lookup("location"))
	viper.BindPFlag(IdKey, rootCmd.PersistentFlags().Lookup("id"))
	viper.BindPFlag(tokenKey, rootCmd.PersistentFlags().Lookup("token"))

	config.Location = viper.GetString(locationKey)
	config.Id = viper.GetUint32(IdKey)
	config.Token = viper.GetString(tokenKey)

	if config.Location == "" {
		if name := singleDevice(); name != "" {
			locationKey = fmt.Sprintf("%s.location", name)
			IdKey = fmt.Sprintf("%s.id", name)
			tokenKey = fmt.Sprintf("%s.token", name)
			config.Location = viper.GetString(locationKey)
			config.Id = viper.GetUint32(IdKey)
			config.Token = viper.GetString(tokenKey)
		}
	}

	return fan2.NewFan2(config.Location, config.Id, config.Token)
}

func getFan() *fan2.Fan {
	fan, err := tryFan()
	exitIfErr(err)
	return fan
}

func singleDevice() string {
	settings := viper.AllSettings()
	var names []string
	for name, val := range settings {
		if name == "_default" {
			continue
		}
		if _, ok := val.(map[string]interface{}); ok {
			names = append(names, name)
		}
	}
	if len(names) == 1 {
		return names[0]
	}
	return ""
}

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	initConfig()

	rootCmd.PersistentFlags().StringVar(&deviceName, "device", "", "use named device")
	rootCmd.PersistentFlags().StringVar(&config.Location, "location", "", "location of the fan")
	rootCmd.PersistentFlags().Uint32Var(&config.Id, "id", 0, "id of the fan")
	rootCmd.PersistentFlags().StringVar(&config.Token, "token", "", "token of the fan")
}
