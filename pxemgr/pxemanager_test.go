package pxemgr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/mayu-infopusher/machinedata"
	"github.com/giantswarm/mayu/hostmgr"
)

const (
	baseConfig = `default_flatcar_version: myversion
network:
  primary_nic:
    ip_range:
      start: 1.1.1.1
      end: 1.1.1.2
templates_env:
  mayu_https_endpoint: https://mayu
`
	configOK  = baseConfig + `  update: "no_updates"`
	configErr = baseConfig + `  update: "update"`
	ignition  = `ignition:
  version: 2.2.0
systemd:
{{if eq  .TemplatesEnv.update "no_updates"}}
  units:
    - name: update-engine.service
      enabled: false
      mask: true{{end}}
`
)

type helper struct {
	dir      string
	fakeEtcd *httptest.Server
	pxeCfg   PXEManagerConfiguration
	req      *http.Request
	w        *httptest.ResponseRecorder
	cluster  *hostmgr.Cluster
}

func setUp(t *testing.T) *helper {
	h := &helper{}

	var err error
	h.dir, err = ioutil.TempDir("", "pxmgr_cloudconfig_")
	if err != nil {
		t.Fatal(err)
	}

	if err := ioutil.WriteFile(filepath.Join(h.dir, "config_ok.yaml"), []byte(configOK), 0644); err != nil { // nolint
		t.Fatal(err)
	}
	if err := ioutil.WriteFile(filepath.Join(h.dir, "config_err.yaml"), []byte(configErr), 0644); err != nil { // nolint
		t.Fatal(err)
	}
	if err := ioutil.WriteFile(filepath.Join(h.dir, "ignition.yaml"), []byte(ignition), 0644); err != nil { // nolint
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(h.dir, "files"), 0644); err != nil {
		t.Fatal(err)
	}

	logger, err := micrologger.New(micrologger.Config{})
	if err != nil {
		t.Fatalf("failed to create logger cluster: %s", err)
	}

	h.cluster, err = hostmgr.NewCluster(h.dir, logger)
	if err != nil {
		t.Fatalf("creating cluster: %s", err)
	}

	h.fakeEtcd = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "{}")
	}))

	h.pxeCfg = PXEManagerConfiguration{
		UseInternalEtcdDiscovery: true,
		NoTLS:                    true,
		// This port is declared only to allow PXEMAnager instantiation (APIPort and
		// PXEPort must be different), the server is not going to be started and we
		// are going to test the handler method directly
		APIPort:        4080,
		FilesDir:       filepath.Join(h.dir, "files"),
		IgnitionConfig: filepath.Join(h.dir, "ignition.yaml"),
		EtcdEndpoint:   h.fakeEtcd.URL,
	}

	hostData := machinedata.HostData{
		Serial: "myserial",
	}
	b := new(bytes.Buffer)
	_ = json.NewEncoder(b).Encode(hostData)
	h.req = httptest.NewRequest("GET", "http://127.0.0.1:4080/ignition?serial=test1234", b)
	h.w = httptest.NewRecorder()

	return h
}

func tearDown(h *helper) {
	os.RemoveAll(h.dir)
	h.fakeEtcd.Close()
}

func TestFinalCloudConfigChecksErrorOk(t *testing.T) {
	h := setUp(t)
	defer tearDown(h)

	h.pxeCfg.ConfigFile = filepath.Join(h.dir, "config_ok.yaml")

	logger, err := micrologger.New(micrologger.Config{})
	if err != nil {
		t.Fatalf("failed to create logger cluster: %s", err)
	}
	h.pxeCfg.Logger = logger

	// instantiate PXEManager (no need to start it)
	mgr, err := PXEManager(h.pxeCfg, h.cluster)
	if err != nil {
		t.Fatalf("unable to create a pxe manager: %s\n", err)
	}

	// call handler func and make assertions on the response recorder
	mgr.ignitionGenerator(h.w, h.req)

	if status := h.w.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	actual := h.w.Body.String()
	expected := `{"ignition":{"config":{},"security":{"tls":{}},"timeouts":{},"version":"2.2.0"},"networkd":{},"passwd":{},"storage":{},"systemd":{"units":[{"enabled":false,"mask":true,"name":"update-engine.service"}]}}
`
	if actual != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			actual, expected)
	}

	// make sure the template is complete
	if !strings.Contains(actual, "update-engine.service") {
		t.Errorf("response body contains incomplete template: %s", actual)
	}
}
