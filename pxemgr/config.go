package pxemgr

import (
	"io/ioutil"
	"os"

	"github.com/giantswarm/mayu/hostmgr"
	"gopkg.in/yaml.v2"
)

func loadConfig(filePath string) (configuration, error) {
	conf := configuration{}

	f, err := os.Open(filePath)
	if err != nil {
		return conf, err
	}
	defer f.Close()

	confBytes, err := ioutil.ReadAll(f)
	if err != nil {
		return conf, err
	}

	err = yaml.Unmarshal(confBytes, &conf)

	return conf, err
}

type configuration struct {
	DefaultCoreOSVersion string `yaml:"default_coreos_version"`
	Network              network
	Profiles             []profile
	TemplatesEnv         map[string]interface{} `yaml:"templates_env"`
}

type profile struct {
	Quantity         int
	Name             string
	Tags             []string
	DisableEngine    bool   `yaml:"disable_engine"`
	CoreOSVersion    string `yaml:"coreos_version"`
	EtcdClusterToken string `yaml:"etcd_cluster_token"`
}

type network struct {
	Interface      string
	BindAddr       string `yaml:"bind_addr"`
	BootstrapRange struct {
		Start string
		End   string
	} `yaml:"bootstrap_range"`
	IPRange struct {
		Start string
		End   string
	} `yaml:"ip_range"`
	Router       string
	DNS          []string
	PXE          bool
	NetworkModel string `yaml:"network_model"`

	IgnoredHosts []string
	StaticHosts  []hostmgr.IPMac
}
