package pxemgr

import (
	"io/ioutil"
	"os"

	"github.com/giantswarm/mayu/hostmgr"
	"gopkg.in/yaml.v2"
	"github.com/giantswarm/microerror"
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

type Network struct {
	PxeInterface     string `yaml:"pxe_interface"`
	MachineInterface string `yaml:"machine_interface"`
	BindAddr         string `yaml:"bind_addr"`
	BootstrapRange   struct {
		Start string
		End   string
	} `yaml:"bootstrap_range"`
	IPRange struct {
		Start string
		End   string
	} `yaml:"ip_range"`
	Router        string
	DNS           []string
	NTP           []string
	PXE           bool
	UEFI          bool
	SubnetSize    string `yaml:"subnet_size"`
	SubnetGateway string `yaml:"subnet_gateway"`
	VlanId        string `yaml:"vlan_id"`
	NetworkModel  string `yaml:"network_model"`
	BondMode      string `yaml:"bond_mode"`

	IgnoredHosts []string
	StaticHosts  []hostmgr.IPMac
}
