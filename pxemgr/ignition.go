package pxemgr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/coreos/ignition/config/types"
	"github.com/golang/glog"
	"gopkg.in/yaml.v2"

	"github.com/giantswarm/mayu/hostmgr"
)

func (mgr *pxeManagerT) WriteIgnitionConfig(host hostmgr.Host, wr io.Writer) error {
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

	ctx.Host.MayuVersion = mgr.version

  ignitionTemplate := mgr.ignitionConfig
  if host.KubernetesSetup {
		for _, metadata := range host.FleetMetadata {
			if strings.Contains(metadata, "core") {
 				ignitionTemplate = "./templates/ignition/k8s_master.yaml"
				break
			}
			ignitionTemplate = "./templates/ignition/k8s_worker.yaml"
		}
	}

	tmpl, err := getTemplate(ignitionTemplate, mgr.templateSnippets)
	if err != nil {
		glog.Fatalln(err)
		return err
	}

	var data bytes.Buffer
	if err = tmpl.Execute(&data, ctx); err != nil {
		glog.Fatalln(err)
		return err
	}
	ignitionJSON, e := convertTemplatetoJSON(data.Bytes(), false)
	if e != nil {
		glog.Fatalln(e)
		return e
	}

	fmt.Fprintln(wr, string(ignitionJSON[:]))
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

			glog.Errorf("Unrecognized keyword: %v", key)
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

func convertTemplatetoJSON(dataIn []byte, pretty bool) ([]byte, error) {
	cfg := types.Config{}

	if err := yaml.Unmarshal(dataIn, &cfg); err != nil {
		return nil, fmt.Errorf("Failed to unmarshal input: %v", err)
	}

	var inCfg interface{}
	if err := yaml.Unmarshal(dataIn, &inCfg); err != nil {
		return nil, fmt.Errorf("Failed to unmarshal input: %v", err)
	}

	if hasUnrecognizedKeys(inCfg, reflect.TypeOf(cfg)) {
		return nil, fmt.Errorf("Unrecognized keys in input, aborting.")
	}

	var (
		dataOut []byte
		err     error
	)

	if pretty {
		dataOut, err = json.MarshalIndent(&cfg, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("Failed to marshal output: %v", err)
		}
		dataOut = append(dataOut, '\n')
	} else {
		dataOut, err = json.Marshal(&cfg)
		if err != nil {
			return nil, fmt.Errorf("Failed to marshal output: %v", err)
		}
	}

	return dataOut, nil
}
