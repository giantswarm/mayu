package pxemgr

import (
	"io/ioutil"
	"os"

	"github.com/giantswarm/mayu/hostmgr"
	"github.com/giantswarm/microerror"
	"gopkg.in/yaml.v2"
)

func LoadConfig(filePath string) (Configuration, error) {
	conf := Configuration{}

	f, err := os.Open(filePath)
	if err != nil {
		return conf, microerror.Mask(err)
	}
	defer f.Close()

	confBytes, err := ioutil.ReadAll(f)
	if err != nil {
		return conf, microerror.Mask(err)
	}

	err = yaml.Unmarshal(confBytes, &conf)

	return conf, microerror.Mask(err)
}

func saveConfig(filePath string, conf Configuration) error {
	confBytes, err := yaml.Marshal(conf)
	ioutil.WriteFile(filePath, confBytes, 0660)
	if err != nil {
		return microerror.Mask(err)
	}
	return nil
}

type Configuration struct {
	DefaultCoreOSVersion string `yaml:"default_coreos_version"`
	Network              Network
	Profiles             []Profile
	TemplatesEnv         map[string]interface{} `yaml:"templates_env"`
}

type Profile struct {
	Quantity         int
	Name             string
	Tags             []string
	DisableEngine    bool   `yaml:"disable_engine"`
	CoreOSVersion    string `yaml:"coreos_version"`
	EtcdClusterToken string `yaml:"etcd_cluster_token"`
}

type NetworkRange struct {
	Start string
	End   string
}

type NetworkRoute struct {
	DestinationCIDR string `yaml:"destination_cidr"`
	RouteHop        string `yaml:"route_hop"`
}

type NetworkModel struct {
	Type               string `yaml:"type"`
	VlanId             string `yaml:"vlan_id"`
	BondMode           string `yaml:"bond_mode"`
	BondInterfaceMatch string `yaml:"bond_interface_match"`
}

type NetworkInterface struct {
	Routes        []NetworkRoute `yaml:"routes"`
	InterfaceName string         `yaml:"interface_name"`
	IPRange       NetworkRange   `yaml:"ip_range"`
	SubnetSize    string         `yaml:"subnet_size"`
	SubnetGateway string         `yaml:"subnet_gateway"`
}

type Network struct {
	BindAddr     string `yaml:"bind_addr"`
	PXE struct {
		Enabled      bool
		PxeInterface NetworkInterface `yaml:"pxe_interface"`
	}

	PrimaryNIC NetworkInterface   `yaml:"primary_nic"`
	ExtraNICs  []NetworkInterface `yaml:"extra_nics"`

	// if set true use UEFI boot, otherwise use legacy BIOS
	UEFI bool

	// DNS and NTP list for installed machines
	DNS []string
	NTP []string

	IgnoredHosts []string
	StaticHosts  []hostmgr.IPMac
}
