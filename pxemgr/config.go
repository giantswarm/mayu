package pxemgr

import (
	"io/ioutil"
	"os"

	"github.com/giantswarm/mayu/hostmgr"
	"gopkg.in/yaml.v2"
)

func LoadConfig(filePath string) (Configuration, error) {
	conf := Configuration{}

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

func saveConfig(filePath string, conf Configuration) error {
	confBytes, err := yaml.Marshal(conf)
	ioutil.WriteFile(filePath, confBytes, 0660)
	if err != nil {
		return err
	}
	return err
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
