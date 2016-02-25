package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/giantswarm/mayu/fs"
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
	if err == nil {
		// hack to make some dnsmasq versions happy
		conf.TFTPRoot, err = filepath.Abs(conf.TFTPRoot)
	}

	conf.filesystem = fs.DefaultFilesystem
	return conf, err

}

type configuration struct {
	filesystem       fs.FileSystem // internal filesystem abstraction to enable testing of file operations.
	FirstStageScript string        `yaml:"first_stage_script"`
	LastStageCC      string        `yaml:"last_stage_cloudconfig"`
	TemplateSnippets string        `yaml:"template_snippets"`
	DNSmasqTmpl      string        `yaml:"dnsmasq_template"`
	TFTPRoot         string
	IPxe             string
	HTTPBindAddr     string `yaml:"http_bind_addr"`
	HTTPPort         int    `yaml:"http_port"`
	NoSecure         bool   `yaml:"no_secure"`
	HTTPSCertFile    string `yaml:"https_cert_file"`
	HTTPSKeyFile     string `yaml:"https_key_file"`
	Dnsmasq          string
	ImagesCacheDir   string                 `yaml:"images_cache_dir"`
	StaticHTMLPath   string                 `yaml:"static_html_path"`
	YochuPath        string                 `yaml:"yochu_path"`
	YochuVersion     string                 `yaml:"yochu_version"`
	TemplatesEnv     map[string]interface{} `yaml:"templates_env"`

	Profiles []profile

	Network network
}

var (
	ErrNotAllCertFilesProvided = errors.New("please configure a key and cert files for TLS secured connections.")
	ErrHTTPSCertFileNotRedable = errors.New("cannot open configured certificate file for TLS secured connections.")
	ErrHTTPSKeyFileNotReadable = errors.New("cannot open configured key file for TLS secured connections.")
)

// Validate checks the configuration based on all Validate* functions
// attached to the configuration struct.
func (c configuration) Validate() (bool, error) {
	if ok, err := c.ValidateHTTPCertificateUsage(); !ok {
		return ok, err
	}

	if ok, err := c.ValidateHTTPCertificateFileExistance(); !ok {
		return ok, err
	}

	return true, nil
}

// ValidateHTTPCertificateUsage checks if the fields HTTPSCertFile and HTTPSKeyFile
// of the configuration struct are set whenever the NoSecure is set to false.
// This makes sure that users are configuring the needed certificate files when
// using TLS encrypted connections.
func (c configuration) ValidateHTTPCertificateUsage() (bool, error) {
	if c.NoSecure == true {
		return true, nil
	}

	if c.NoSecure == false && c.HTTPSCertFile != "" && c.HTTPSKeyFile != "" {
		return true, nil
	}

	return false, ErrNotAllCertFilesProvided
}

// ValidateHTTPCertificateFileExistance checks if the filenames configured
// in the fields HTTPSCertFile and HTTPSKeyFile can be stat'ed to make sure
// they actually exist.
func (c configuration) ValidateHTTPCertificateFileExistance() (bool, error) {
	if c.NoSecure == true {
		return true, nil
	}

	if _, err := c.filesystem.Stat(c.HTTPSCertFile); err != nil {
		return false, ErrHTTPSCertFileNotRedable
	}

	if _, err := c.filesystem.Stat(c.HTTPSKeyFile); err != nil {
		return false, ErrHTTPSKeyFileNotReadable
	}

	return true, nil
}

type profile struct {
	Quantity int
	Name     string
	Tags     []string
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

func thisHost() string {
	scheme := "https"
	if conf.NoSecure {
		scheme = "http"
	}

	return fmt.Sprintf("%s://%s:%d", scheme, conf.Network.BindAddr, conf.HTTPPort)
}
