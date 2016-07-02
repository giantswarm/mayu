package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/giantswarm/mayu/hostmgr"
	"github.com/ryanuber/columnize"
	"github.com/spf13/cobra"
)

const (
	defaultListFields = "ip,serial,profile,ipmiaddr,providerid,etcdtoken,metadata,coreos,state,lastboot"
	timestampFormat   = "2006-01-02 15:04:05"
)

type ListFlags struct {
	Fields   string
	NoLegend bool
}

type hostToField func(host *hostmgr.Host) string

var (
	listCmd = &cobra.Command{
		Use:   "list [--no-legend] [--fields]",
		Short: "List machines.",
		Long: `List the state of all machines in a Mayu datacenter.

For easily parsable output, you can remove the column headers:
	mayuctl list --no-legend

Or, choose the columns to display:
	mayuctl list --fields=ip,machineid,ipmiaddr`,
		Run: listRun,
	}

	listFields = map[string]hostToField{
		"ip": func(host *hostmgr.Host) string {
			if host.InternalAddr == nil {
				return "-"
			}
			return host.InternalAddr.String()
		},
		"serial": func(host *hostmgr.Host) string {
			return host.Serial
		},
		"profile": func(host *hostmgr.Host) string {
			if host.Profile == "" {
				return "-"
			}
			return host.Profile
		},
		"ipmiaddr": func(host *hostmgr.Host) string {
			if host.IPMIAddr == nil {
				return "-"
			}
			return host.IPMIAddr.String()
		},
		"providerid": func(host *hostmgr.Host) string {
			if host.ProviderId == "" {
				return "-"
			}
			return host.ProviderId
		},
		"etcdtoken": func(host *hostmgr.Host) string {
			return host.EtcdClusterToken
		},
		"metadata": func(host *hostmgr.Host) string {
			return host.FleetMetadata.String()
		},
		"coreos": func(host *hostmgr.Host) string {
			return host.CoreOSVersion
		},
		"yochu": func(host *hostmgr.Host) string {
			return host.YochuVersion
		},
		"fleet": func(host *hostmgr.Host) string {
			return host.FleetVersion
		},
		"etcd": func(host *hostmgr.Host) string {
			return host.EtcdVersion
		},
		"docker": func(host *hostmgr.Host) string {
			return host.DockerVersion
		},
		"machineid": func(host *hostmgr.Host) string {
			return host.MachineID
		},
		"state": func(host *hostmgr.Host) string {
			return hostmgr.HostStateMap()[host.State]
		},
		"lastboot": func(host *hostmgr.Host) string {
			return host.LastBoot.Format(timestampFormat)
		},
	}

	listFlags = &ListFlags{}
)

func init() {
	listCmd.PersistentFlags().BoolVar(&listFlags.NoLegend, "no-legend", false, "Do not print a legend (column headers)")
	listCmd.PersistentFlags().StringVar(&listFlags.Fields, "fields", defaultListFields, fmt.Sprintf("Columns to print for each Machine. Valid fields are %q", strings.Join(hostToFieldKeys(listFields), ",")))
}

func listRun(cmd *cobra.Command, args []string) {
	if listFlags.Fields == "" {
		fmt.Printf("Invalid fields parameter. Please choose valid fields: %s\n", strings.Join(hostToFieldKeys(listFields), ","))
		os.Exit(1)
	}

	cols := strings.Split(listFlags.Fields, ",")
	for _, s := range cols {
		if _, ok := listFields[s]; !ok {
			fmt.Printf("Invalid field: %q.\n\nUsage: %s\n", s, cmd.Usage())
			os.Exit(1)
		}
	}

	hosts, err := mayu.List()
	assert(err)

	lines := []string{}
	for _, host := range hosts {
		var f []string
		for _, c := range cols {
			f = append(f, listFields[c](&host))
		}
		lines = append(lines, strings.Join(f, "|"))
	}
	sort.Strings(lines)

	if !listFlags.NoLegend {
		lines = append([]string{strings.ToUpper(strings.Join(cols, "|"))}, lines...)
	}

	fmt.Println(columnize.SimpleFormat(lines))
}

func hostToFieldKeys(m map[string]hostToField) (keys []string) {
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return
}
