package main

import (
	"github.com/spf13/cobra"
)

var (
	reconfigureCmd = &cobra.Command{
		Use:   "reconfigure",
		Short: "Reinstall the hosts with state 'configured|installing|unknown'.",
		Long:  `Reinstall the hosts with state 'configured|installing|unknown'.`,
		Run:   reconfigureRun,
	}
)

func reconfigureRun(cmd *cobra.Command, args []string) {
	err := mayu.Reconfigure()
	assert(err)
}
