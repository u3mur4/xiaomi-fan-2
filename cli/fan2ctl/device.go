package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	var deviceCmd = &cobra.Command{
		Use:     "device",
		Short:   "manage device configurations",
		GroupID: "config",
	}

	var addCmd = &cobra.Command{
		Use:   "add <name>",
		Args:  cobra.ExactArgs(1),
		Short: "add or update a device",
		Run: func(cmd *cobra.Command, args []string) {
			if config.Id == 0 || config.Location == "" || config.Token == "" {
				fmt.Fprintln(os.Stderr, "error: --id, --location and --token are required")
				os.Exit(-1)
			}

			name := args[0]
			viper.Set(fmt.Sprintf("%s.location", name), config.Location)
			viper.Set(fmt.Sprintf("%s.id", name), config.Id)
			viper.Set(fmt.Sprintf("%s.token", name), config.Token)

			cfgFile := viper.ConfigFileUsed()
			err := viper.WriteConfigAs(cfgFile)
			exitIfErr(err)
			fmt.Println("saved device:", name)
		},
	}

	var listCmd = &cobra.Command{
		Use:   "list",
		Short: "list saved devices",
		Run: func(cmd *cobra.Command, args []string) {
			settings := viper.AllSettings()
			defaultDevice, _ := settings["_default"].(string)
			var names []string
			for name, val := range settings {
				if name == "_default" {
					continue
				}
				if _, ok := val.(map[string]any); ok {
					names = append(names, name)
				}
			}
			sort.Strings(names)
			for _, name := range names {
				mark := " "
				if name == defaultDevice {
					mark = "*"
				}
				fmt.Printf("  %s %s\n", mark, name)
			}
		},
	}

	var removeCmd = &cobra.Command{
		Use:   "remove <name>",
		Args:  cobra.ExactArgs(1),
		Short: "remove a device",
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			viper.Set(fmt.Sprintf("%s.location", name), nil)
			viper.Set(fmt.Sprintf("%s.id", name), nil)
			viper.Set(fmt.Sprintf("%s.token", name), nil)

			if viper.GetString("_default") == name {
				viper.Set("_default", nil)
			}

			cfgFile := viper.ConfigFileUsed()
			err := viper.WriteConfigAs(cfgFile)
			exitIfErr(err)
			fmt.Println("removed device:", name)
		},
	}

	var defaultCmd = &cobra.Command{
		Use:   "default [<name>]",
		Short: "show or set the default device",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) > 0 {
				viper.Set("_default", args[0])
				cfgFile := viper.ConfigFileUsed()
				err := viper.WriteConfigAs(cfgFile)
				exitIfErr(err)
				fmt.Println("default device set to:", args[0])
				return
			}
			def := viper.GetString("_default")
			if def == "" {
				fmt.Println("no default device set")
			} else {
				fmt.Println(def)
			}
		},
	}

	deviceCmd.AddCommand(addCmd, listCmd, removeCmd, defaultCmd)
	rootCmd.AddCommand(deviceCmd)
}
