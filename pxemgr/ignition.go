package pxemgr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"

	"github.com/coreos/ignition/config/types"
	"github.com/golang/glog"
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

	ctx.Host.MayuVersion = mgr.version

	tmpl, err := getTemplate(mgr.ignitionConfig, mgr.templateSnippets)
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
