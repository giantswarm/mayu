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

	if err := ioutil.WriteFile(filepath.Join(h.dir, "config_ok.yaml"), []byte(configOK), 0644); err != nil {
		t.Fatal(err)
	}
	if err := ioutil.WriteFile(filepath.Join(h.dir, "config_err.yaml"), []byte(configErr), 0644); err != nil {
		t.Fatal(err)
	}
	if err := ioutil.WriteFile(filepath.Join(h.dir, "last_stage_cloudconfig.yaml"), []byte(lastStageCloudconfig), 0644); err != nil {
		t.Fatal(err)
	}

	h.cluster, err = hostmgr.NewCluster(h.dir, true)
	if err != nil {
		t.Fatalf("creating cluster: %s", err)
	}

	h.fakeEtcd = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "{}")
	}))

	h.pxeCfg = PXEManagerConfiguration{
		UseInternalEtcdDiscovery: true,
		NoTLS:                true,
		APIPort:              4080,
		PXEPort:              4081,
		LastStageCloudconfig: filepath.Join(h.dir, "last_stage_cloudconfig.yaml"),
		EtcdEndpoint:         h.fakeEtcd.URL,
	}

	hostData := machinedata.HostData{
		Serial: "myserial",
	}
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(hostData)
	h.req = httptest.NewRequest("POST", "http://127.0.0.1:4080/final-cloud-config.yaml", b)
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
	mgr, err := PXEManager(h.pxeCfg, h.cluster)
	if err != nil {
		t.Fatalf("unable to create a pxe manager: %s\n", err)
	}
	mgr.configGenerator(h.w, h.req)

	if status := h.w.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	actual := h.w.Body.String()
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
	h := setUp(t)
	defer tearDown(h)

	h.pxeCfg.ConfigFile = filepath.Join(h.dir, "config_err.yaml")
	mgr, err := PXEManager(h.pxeCfg, h.cluster)
	if err != nil {
		t.Fatalf("unable to create a pxe manager: %s\n", err)
	}
	mgr.configGenerator(h.w, h.req)

	if status := h.w.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusInternalServerError)
	}

	actual := h.w.Body.String()
	expected := `generating final stage cloudConfig failed: template: last_stage_cloudconfig.yaml`
	if !strings.Contains(actual, expected) {
		t.Errorf("handler returned unexpected body: got %v want %v",
			actual, expected)
	}

	// make sure we don't get partial templates
	if strings.Contains(actual, "before_key: ok") {
		t.Errorf("response body contains partial template: %s", actual)
	}
}
