package pxemgr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
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

func join(s ...string) string {
	// first arg is sep, remaining args are strings to join
	return strings.Join(s[1:], s[0])
}

func getTemplate(path, snippets string) (*template.Template, error) {
	maybeInitSnippets(snippets)
	templates := []string{path}
	templates = append(templates, snippetsFiles...)

	tmpl, err := template.ParseFiles(templates...)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	tmpl.Funcs(map[string]interface{}{
		"join": join,
	})

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
