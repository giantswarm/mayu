package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	setCmd = &cobra.Command{
		Use:   "set <serial> <key> <value>",
		Short: "Set metadata of machines (metadata, providerid, ipmiaddr, cabinet, state, etcdtoken).",
		Long:  "Set metadata of machines (metadata, providerid, ipmiaddr, cabinet, state, etcdtoken).",
		Run:   setRun,
	}
)

func setRun(cmd *cobra.Command, args []string) {
	if len(args) != 3 {
		fmt.Printf("Usage: %s\n", cmd.Usage())
		os.Exit(1)
	}

	switch args[1] {
	case "metadata":
		err := mayu.SetMetadata(args[0], args[2])
		assert(err)
	case "providerid":
		err := mayu.SetProviderId(args[0], args[2])
		assert(err)
	case "ipmiaddr":
		err := mayu.SetIPMIAddr(args[0], args[2])
		assert(err)
	case "cabinet":
		err := mayu.SetCabinet(args[0], args[2])
		assert(err)
	case "state":
		err := mayu.SetState(args[0], args[2])
		assert(err)
	case "etcdtoken":
		err := mayu.SetEtcdClusterToken(args[0], args[2])
		assert(err)
	default:
		fmt.Printf("setting key with name '%s' is not supported\n", args[1])
		os.Exit(1)
	}
}
