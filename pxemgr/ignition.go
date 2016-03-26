package pxemgr

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"

	"github.com/coreos/ignition/config"
	"gopkg.in/yaml.v2"
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

func (mgr *pxeManagerT) WriteIgnitionConfig(host hostmgr.Host, wr io.Writer) error {
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
		EtcdDiscoveryUrl: mgr.cluster.Config.EtcdDiscoveryURL,
		MayuHost:         mgr.config.Network.BindAddr,
		MayuPort:         mgr.httpPort,
		MayuURL:          mgr.thisHost(),
		PostBootURL:      mgr.thisHost() + "/admin/host/" + host.Serial + "/boot_complete",
		NoTLS:            mgr.noTLS,
		TemplatesEnv:     mgr.config.TemplatesEnv,
	}

	tmpl, err := getTemplate(mgr.ignitionConfig, mgr.templateSnippets)
	if err != nil {
		glog.Fatalln(err)
		return err
	}

	if err = tmpl.Execute(wr, ctx); err != nil {
		glog.Fatalln(err)
		return err
	}
	ignitionJSON, e := convertTemplatetoJSON(wr, false)
	if e != nil {
		glog.Fatalln(e)
		return e
	}
	glog.Info(ignitionJSON)

	return nil
}

// hasUnrecognizedKeys finds unrecognized keys and warns about them on stderr.
// returns false when no unrecognized keys were found, true otherwise.
func hasUnrecognizedKeys(inCfg interface{}, refType reflect.Type) (warnings bool) {
	if refType.Kind() == reflect.Ptr {
		refType = refType.Elem()
	}
	switch inCfg.(type) {
	case map[interface{}]interface{}:
		ks := inCfg.(map[interface{}]interface{})
	keys:
		for key := range ks {
			for i := 0; i < refType.NumField(); i++ {
				sf := refType.Field(i)
				tv := sf.Tag.Get("yaml")
				if tv == key {
					if warn := hasUnrecognizedKeys(ks[key], sf.Type); warn {
						warnings = true
					}
					continue keys
				}
			}

			stderr("Unrecognized keyword: %v", key)
			warnings = true
		}
	case []interface{}:
		ks := inCfg.([]interface{})
		for i := range ks {
			if warn := hasUnrecognizedKeys(ks[i], refType.Elem()); warn {
				warnings = true
			}
		}
	default:
	}
	return
}

func convertTemplatetoJSON(dataIn string, pretty bool) (string, error) {
	cfg := config.Config{}
	//dataIn, err := ioutil.ReadFile(filePath)
	//if err != nil {
	//	return "", fmt.Errorf("Failed to read: %v", err)
	//}

	if err := yaml.Unmarshal(dataIn, &cfg); err != nil {
		return "", fmt.Errorf("Failed to unmarshal input: %v", err)
	}

	var inCfg interface{}
	if err := yaml.Unmarshal(dataIn, &inCfg); err != nil {
		return "", fmt.Errorf("Failed to unmarshal input: %v", err)
	}

	if hasUnrecognizedKeys(inCfg, reflect.TypeOf(cfg)) {
		return "", fmt.Errorf("Unrecognized keys in input, aborting.")
	}

	var dataOut []byte
	if pretty {
		dataOut, err = json.MarshalIndent(&cfg, "", "  ")
		dataOut = append(dataOut, '\n')
	} else {
		dataOut, err = json.Marshal(&cfg)
	}
	if err != nil {
		return "", fmt.Errorf("Failed to marshal output: %v", err)
	}

	return dataOut, nil
}
