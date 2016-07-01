package pxemgr

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"

	"github.com/golang/glog"
	"github.com/gorilla/mux"
)

type EtcdNode struct {
	Key   string      `json:"key"`
	Value string      `json:"value,omitempty"`
	Nodes []*EtcdNode `json:"nodes,omitempty"`
	Dir   bool        `json:"dir,omitempty"`
}

type EtcdResponse struct {
	Action string    `json:"action"`
	Node   *EtcdNode `json:"node,omitempty"`
}

type EtcdResponseError struct {
	ErrorCode int    `json:"errorCode"`
	Message   string `json:"message"`
	Cause     string `json:"cause"`
}

func (mgr *pxeManagerT) defineEtcdDiscoveryRoutes(etcdRouter *mux.Router) {
	etcdRouter.PathPrefix("/new").Methods("PUT").HandlerFunc(mgr.etcdDiscoveryNewCluster)

	tokenRouter := etcdRouter.PathPrefix("/{token:[a-f0-9]{32}}").Subrouter()
	tokenRouter.PathPrefix("/_config/size").Methods("GET").HandlerFunc(mgr.etcdDiscoveryProxyHandler)
	tokenRouter.PathPrefix("/_config/size").Methods("PUT").HandlerFunc(mgr.etcdDiscoveryProxyHandler)
	tokenRouter.PathPrefix("/{machine}").Methods("PUT").HandlerFunc(mgr.etcdDiscoveryProxyHandler)
	tokenRouter.PathPrefix("/{machine}").Methods("GET").HandlerFunc(mgr.etcdDiscoveryProxyHandler)
	tokenRouter.PathPrefix("/{machine}").Methods("DELETE").HandlerFunc(mgr.etcdDiscoveryProxyHandler)
	tokenRouter.Methods("GET").HandlerFunc(mgr.etcdDiscoveryProxyHandler)

	etcdRouter.Methods("GET").HandlerFunc(mgr.etcdDiscoveryHandler)
}

func (mgr *pxeManagerT) etcdDiscoveryHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r,
		"https://github.com/giantswarm/mayu/blob/master/docs/etcd-discovery.md",
		http.StatusMovedPermanently,
	)
}

func (mgr *pxeManagerT) etcdDiscoveryNewCluster(w http.ResponseWriter, r *http.Request) {
	var err error
	size := mgr.defaultEtcdQuorumSize
	s := r.FormValue("size")
	if s != "" {
		size, err = strconv.Atoi(s)
		if err != nil {
			httpError(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	token, err := mgr.cluster.GenerateEtcdDiscoveryToken(mgr.etcdEndpoint, size)
	if err != nil {
		httpError(w, fmt.Sprintf("Unable to generate token '%v'", err), 400)
		return
	}

	glog.V(2).Infof("New cluster created '%s'", token)

	fmt.Fprintf(w, "%s/%s", mgr.etcdDiscoveryBaseURL(), token)
}

func (mgr *pxeManagerT) etcdDiscoveryBaseURL() string {
	return fmt.Sprintf("%s/etcd", mgr.thisHost())
}

func (mgr *pxeManagerT) etcdDiscoveryProxyHandler(w http.ResponseWriter, r *http.Request) {
	resp, err := mgr.etcdDiscoveryProxyRequest(r)
	if err != nil {
		httpError(w, fmt.Sprintf("Error proxying request to etcd '%v'", err), 500)
	}

	copyHeader(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func (mgr *pxeManagerT) etcdDiscoveryProxyRequest(r *http.Request) (*http.Response, error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	for i := 0; i <= 10; i++ {
		u := url.URL{
			Scheme:   "http",
			Host:     mgr.etcdEndpoint,
			Path:     path.Join("v2", "keys", "_etcd", "registry", strings.TrimPrefix(r.URL.Path, "/etcd")),
			RawQuery: r.URL.RawQuery,
		}

		buf := bytes.NewBuffer(body)
		glog.V(2).Infof("Body '%s'", body)

		outreq, err := http.NewRequest(r.Method, u.String(), buf)
		if err != nil {
			return nil, err
		}

		copyHeader(outreq.Header, r.Header)

		client := http.Client{}
		resp, err := client.Do(outreq)
		if err != nil {
			return nil, err
		}

		return resp, nil
	}

	return nil, errors.New("All attempts at proxying to etcd failed")
}

// copyHeader copies all of the headers from dst to src.
func copyHeader(dst, src http.Header) {
	for k, v := range src {
		for _, q := range v {
			dst.Add(k, q)
		}
	}
}
