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

	"github.com/giantswarm/mayu-infopusher/machinedata"
	"github.com/giantswarm/mayu/hostmgr"
)

const (
	baseConfig = `default_coreos_version: myversion
network:
  ip_range:
    start: 1.1.1.1
    end: 1.1.1.2
templates_env:
  mayu_https_endpoint: https://mayu
`
	configOK             = baseConfig + "  storage: mystorage"
	configErr            = baseConfig
	lastStageCloudconfig = `before_key: ok
{{if eq .TemplatesEnv.storage "mystorage"}}storage: mystorage{{end}}
after_key: ok
`
)

var (
	dir      string
	fakeEtcd *httptest.Server
	pxeCfg   PXEManagerConfiguration
	req      *http.Request
	w        *httptest.ResponseRecorder
	cluster  *hostmgr.Cluster
)

func setUp(t *testing.T) {
	var err error
	dir, err = ioutil.TempDir("", "pxmgr_cloudconfig_")
	if err != nil {
		t.Fatal(err)
	}

	if err := ioutil.WriteFile(filepath.Join(dir, "config_ok.yaml"), []byte(configOK), 0644); err != nil {
		t.Fatal(err)
	}
	if err := ioutil.WriteFile(filepath.Join(dir, "config_err.yaml"), []byte(configErr), 0644); err != nil {
		t.Fatal(err)
	}
	if err := ioutil.WriteFile(filepath.Join(dir, "last_stage_cloudconfig.yaml"), []byte(lastStageCloudconfig), 0644); err != nil {
		t.Fatal(err)
	}

	cluster, err = hostmgr.NewCluster(dir, true)
	if err != nil {
		t.Fatalf("creating cluster: %s", err)
	}

	fakeEtcd = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "{}")
	}))

	pxeCfg = PXEManagerConfiguration{
		UseInternalEtcdDiscovery: true,
		NoTLS:                true,
		APIPort:              4080,
		PXEPort:              4081,
		BindAddress:          "0.0.0.0",
		UseIgnition:          false,
		LastStageCloudconfig: filepath.Join(dir, "last_stage_cloudconfig.yaml"),
		Version:              "1.0.0",
		EtcdEndpoint:         fakeEtcd.URL,
	}

	hostData := machinedata.HostData{
		Serial: "myserial",
	}
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(hostData)
	req = httptest.NewRequest("POST", "http://127.0.0.1:4080/final-cloud-config.yaml", b)
	w = httptest.NewRecorder()
}

func tearDown() {
	os.RemoveAll(dir)
	fakeEtcd.Close()
}

func TestFinalCloudConfigChecksErrorOk(t *testing.T) {
	setUp(t)
	defer tearDown()

	pxeCfg.ConfigFile = filepath.Join(dir, "config_ok.yaml")
	mgr, err := PXEManager(pxeCfg, cluster)
	if err != nil {
		t.Fatalf("unable to create a pxe manager: %s\n", err)
	}
	mgr.configGenerator(w, req)

	if status := w.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	actual := w.Body.String()
	expected := `before_key: ok
storage: mystorage
after_key: ok
`
	if actual != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			actual, expected)
	}

	// make sure the template is complete
	if !strings.Contains(actual, "after_key: ok") {
		t.Errorf("response body contains incomplete template: %s", actual)
	}
}

func TestFinalCloudConfigChecksErrorFail(t *testing.T) {
	setUp(t)
	defer tearDown()

	pxeCfg.ConfigFile = filepath.Join(dir, "config_err.yaml")
	mgr, err := PXEManager(pxeCfg, cluster)
	if err != nil {
		t.Fatalf("unable to create a pxe manager: %s\n", err)
	}
	mgr.configGenerator(w, req)

	if status := w.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusInternalServerError)
	}

	actual := w.Body.String()
	expected := `generating final stage cloudConfig failed: template: last_stage_cloudconfig.yaml:2:5: executing "last_stage_cloudconfig.yaml" at <eq .TemplatesEnv.sto...>: error calling eq: invalid type for comparison`
	if actual != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			actual, expected)
	}

	// make sure we don't get partial templates
	if strings.Contains(actual, "before_key: ok") {
		t.Errorf("response body contains partial template: %s", actual)
	}
}
