package main

import (
	"fmt"
	"github.com/giantswarm/mayu/pxemgr"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
)

type ConfigFlags struct {
	Method     string
	ConfigFile string
}

var (
	configCmd = &cobra.Command{
		Use:   "config get",
		Short: "fetch or change curent mayu config",
		Long:  "fetch or change curent mayu config",
		Run:   configRun,
	}
	configFlags = &ConfigFlags{}
)

func init() {
	configCmd.PersistentFlags().StringVar(&configFlags.Method, "method", "get", "Method - 'get' or 'set', default is get. Set method requires config-file flag.")
	configCmd.PersistentFlags().StringVar(&configFlags.ConfigFile, "config-file", "", "File path to configfile, required when method is 'set', optional for 'get' (response is written to file)")
}

func configRun(cmd *cobra.Command, args []string) {
	if configFlags.Method == "get" {
		config, err := mayu.GetConfig()
		if err != nil {
			println("ERROR: when fetching config from mayu", err)
			os.Exit(1)
		}
		if configFlags.ConfigFile != "" {
			err := ioutil.WriteFile(configFlags.ConfigFile, []byte(config), 0660)
			if err != nil {
				println("ERROR: when saving config to file ", err)
			}
		} else {
			fmt.Println(config)
		}
	} else if configFlags.Method == "set" {
		if configFlags.ConfigFile == "" {
			println("ERROR: method set but no config-file passed")
			os.Exit(1)
		}
		conf, err := pxemgr.LoadConfig(configFlags.ConfigFile)
		if err != nil {
			println("ERROR: parsing config file: ", err)
		}
		err = mayu.SetConfig(conf)
		if err != nil {
			println("ERROR: when setting config in mayu", err)
			os.Exit(1)
		}
	} else {
		fmt.Printf("ERROR: unknow operation %s", configFlags.Method)
		os.Exit(1)
	}
}
