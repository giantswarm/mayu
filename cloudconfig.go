package main

import (
	"io"
	"io/ioutil"
	"os"
	"path"
	"text/template"

	"gopkg.in/yaml.v2"

	"github.com/giantswarm/mayu/hostmgr"
	"github.com/golang/glog"
)

var snippetsFiles []string

func maybeInitSnippets() {
	if snippetsFiles != nil {
		return
	}
	snippetsFiles = []string{}

	if len(globalFlags.templateSnippets) > 0 {
		if _, err := os.Stat(globalFlags.templateSnippets); err == nil {
			if fis, err := ioutil.ReadDir(globalFlags.templateSnippets); err == nil {
				for _, fi := range fis {
					snippetsFiles = append(snippetsFiles, path.Join(globalFlags.templateSnippets, fi.Name()))
				}
			}
		}
	}
}

func getTemplate(path string) (*template.Template, error) {
	maybeInitSnippets()
	templates := []string{path}
	templates = append(templates, snippetsFiles...)
	glog.V(10).Infof("templates: %+v\n", templates)

	return template.ParseFiles(templates...)
}

func validateCC(cloudConfig []byte) error {
	cc := map[string]interface{}{}
	return yaml.Unmarshal(cloudConfig, &cc)
}

func (mgr *pxeManagerT) writeLastStageCC(host hostmgr.Host, wr io.Writer) error {
	ctx := struct {
		Host             hostmgr.Host
		EtcdDiscoveryUrl string
		ClusterNetwork   network
		MayuHost         string
		MayuPort         string
		MayuURL          string
		PostBootURL      string
		NoSecure         bool
		TemplatesEnv     map[string]interface{}
	}{
		Host:             host,
		ClusterNetwork:   conf.Network,
		EtcdDiscoveryUrl: mgr.cluster.Config.EtcdDiscoveryURL,
		MayuHost:         conf.Network.BindAddr,
		MayuPort:         globalFlags.httpPort,
		MayuURL:          thisHost(),
		PostBootURL:      thisHost() + "/admin/host/" + host.Serial + "/boot_complete",
		NoSecure:         globalFlags.noSecure,
		TemplatesEnv:     conf.TemplatesEnv,
	}

	tmpl, err := getTemplate(globalFlags.lastStageCloudconfig)
	if err != nil {
		glog.Fatalln(err)
		return err
	}

	return tmpl.Execute(wr, ctx)
}
