package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	overrideCmd = &cobra.Command{
		Use:   "override <serial> <property> <value>",
		Short: "Overrides globally defined properties for a host: EtcdDiscoveryURL, docker_version, yochu_version, etc",
		Long:  "Overrides globally defined properties for a host: EtcdDiscoveryURL, docker_version, yochu_version, etc",
		Run:   overrideRun,
	}
)

func overrideRun(cmd *cobra.Command, args []string) {
	if len(args) != 3 {
		fmt.Printf("Usage: %s\n", cmd.Usage())
		os.Exit(1)
	}

	mayu.Override(args[0], args[1], args[2])
}
