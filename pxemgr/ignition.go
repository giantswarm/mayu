package pxemgr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/giantswarm/microerror"
	"gopkg.in/yaml.v2"

	"github.com/giantswarm/mayu/hostmgr"
)

func (mgr *pxeManagerT) WriteIgnitionConfig(host hostmgr.Host, wr io.Writer) error {
	etcdClusterToken := mgr.cluster.Config.DefaultEtcdClusterToken

	if host.EtcdClusterToken != "" {
		etcdClusterToken = host.EtcdClusterToken
	}

	mergedTemplatesEnv := mgr.config.TemplatesEnv
	for k, v := range host.Overrides {
		mergedTemplatesEnv[k] = v
	}

	ctx := struct {
		Host             hostmgr.Host
		EtcdDiscoveryUrl string
		ClusterNetwork   Network
		MayuHost         string
		MayuPort         int
		MayuURL          string
		PostBootURL      string
		NoTLS            bool
		TemplatesEnv     map[string]interface{}
		Files            Files
	}{
		Host:             host,
		ClusterNetwork:   mgr.config.Network,
		EtcdDiscoveryUrl: fmt.Sprintf("%s/%s", mgr.etcdDiscoveryUrl, etcdClusterToken),
		MayuHost:         mgr.config.Network.BindAddr,
		MayuPort:         mgr.apiPort,
		MayuURL:          mgr.apiURL(),
		PostBootURL:      mgr.apiURL() + "/admin/host/" + host.Serial + "/boot_complete",
		NoTLS:            mgr.noTLS,
		TemplatesEnv:     mergedTemplatesEnv,
	}

	files, err := mgr.RenderFiles(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	ctx.Files = *files
	tmpl, err := getTemplate(mgr.ignitionConfig, mgr.templateSnippets)
	if err != nil {
		return microerror.Mask(err)
	}

	var data bytes.Buffer
	if err = tmpl.Execute(&data, ctx); err != nil {
		return microerror.Mask(err)
	}
	ignitionJSON, err := convertTemplateToJSON(data.Bytes(), false)
	if err != nil {
		return microerror.Mask(err)
	}
	host.State = hostmgr.Installing
	fmt.Fprintln(wr, string(ignitionJSON[:]))
	return nil
}

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

func join(sep string, i []interface{}) string {
	var s []string
	for _, si := range i {
		s = append(s, si.(string))
	}
	return strings.Join(s, sep)
}

func getTemplate(path, snippets string) (*template.Template, error) {
	maybeInitSnippets(snippets)
	templates := []string{path}
	templates = append(templates, snippetsFiles...)

	name := filepath.Base(path)
	tmpl := template.New(name)
	tmpl.Funcs(map[string]interface{}{
		"join": join,
	})

	var err error
	tmpl, err = tmpl.ParseFiles(templates...)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return tmpl, nil
}

func convertTemplateToJSON(dataIn []byte, pretty bool) ([]byte, error) {
	cfg := Config{}

	if err := yaml.Unmarshal(dataIn, &cfg); err != nil {
		return nil, microerror.Maskf(executionFailedError, "failed to unmarshal input: %v", err)
	}

	var (
		dataOut []byte
		err     error
	)

	if pretty {
		dataOut, err = json.MarshalIndent(&cfg, "", "  ")
		if err != nil {
			return nil, microerror.Maskf(executionFailedError, "failed to marshal output: %v", err)
		}
		dataOut = append(dataOut, '\n')
	} else {
		dataOut, err = json.Marshal(&cfg)
		if err != nil {
			return nil, microerror.Maskf(executionFailedError, "failed to marshal output: %v", err)
		}
	}

	return dataOut, nil
}
