package pxemgr

import (
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/giantswarm/mayu/hostmgr"
	"github.com/golang/glog"
)

var snippetsFiles []string

func maybeInitSnippets(snippets string) {
	if snippetsFiles != nil {
		return
	}
	snippetsFiles = []string{}

	if len(snippets) > 0 {
		if _, err := os.Stat(snippets); err == nil {
			if fis, err := ioutil.ReadDir(snippets); err == nil {
				for _, fi := range fis {
					snippetsFiles = append(snippetsFiles, path.Join(snippets, fi.Name()))
				}
			}
		}
	}
}

func getTemplate(path, snippets string) (*template.Template, error) {
	maybeInitSnippets(snippets)
	templates := []string{path}
	templates = append(templates, snippetsFiles...)
	glog.V(10).Infof("templates: %+v\n", templates)

	return template.ParseFiles(templates...)
}

func (mgr *pxeManagerT) WriteLastStageCC(host hostmgr.Host, wr io.Writer) error {
	etcdDiscoveryURL := mgr.cluster.Config.EtcdDiscoveryURL

	if hostDiscoveryURL, exists := host.Overrides["EtcdDiscoveryURL"]; exists {
		etcdDiscoveryURL = hostDiscoveryURL.(string)
	}

	mergedTemplatesEnv := mgr.config.TemplatesEnv
	for k, v := range host.Overrides {
		if k != "EtcdDiscoveryUrl" {
			mergedTemplatesEnv[k] = v
		}
	}

	ctx := struct {
		Host             hostmgr.Host
		EtcdDiscoveryUrl string
		ClusterNetwork   network
		MayuHost         string
		MayuPort         int
		MayuURL          string
		PostBootURL      string
		NoTLS            bool
		TemplatesEnv     map[string]interface{}
	}{
		Host:             host,
		ClusterNetwork:   mgr.config.Network,
		EtcdDiscoveryUrl: etcdDiscoveryURL,
		MayuHost:         mgr.config.Network.BindAddr,
		MayuPort:         mgr.httpPort,
		MayuURL:          mgr.thisHost(),
		PostBootURL:      mgr.thisHost() + "/admin/host/" + host.Serial + "/boot_complete",
		NoTLS:            mgr.noTLS,
		TemplatesEnv:     mergedTemplatesEnv,
	}

	cloudConfigTemplate := mgr.lastStageCloudconfig
	if host.KubernetesSetup {
		for _, metadata := range host.FleetMetadata {
			if strings.Contains(metadata, "core") {
				cloudConfigTemplate = "./templates/last_stage_k8s_master_cloudconfig.yaml"
				break
			} else {
				cloudConfigTemplate = "./templates/last_stage_k8s_worker_cloudconfig.yaml"
			}
		}
	}

	tmpl, err := getTemplate(cloudConfigTemplate, mgr.templateSnippets)
	if err != nil {
		glog.Fatalln(err)
		return err
	}

	return tmpl.Execute(wr, ctx)
}
