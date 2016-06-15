package pxemgr

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/giantswarm/mayu/hostmgr"
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
	tokenRouter.PathPrefix("/_config/size").Methods("GET").HandlerFunc(mgr.etcdDiscoveryTokenSize)
	tokenRouter.PathPrefix("/_config/size").Methods("PUT").HandlerFunc(mgr.etcdDiscoveryUpdateTokenSize)
	tokenRouter.PathPrefix("/{machine}").Methods("PUT").HandlerFunc(mgr.etcdDiscoveryAddMachine)
	tokenRouter.PathPrefix("/{machine}").Methods("GET").HandlerFunc(mgr.etcdDiscoveryMachine)
	tokenRouter.PathPrefix("/{machine}").Methods("DELETE").HandlerFunc(mgr.etcdDiscoveryDeleteMachine)
	tokenRouter.Methods("GET").HandlerFunc(mgr.etcdDiscoveryToken)

	etcdRouter.Methods("GET").HandlerFunc(mgr.etcdDiscovery)
}

func (mgr *pxeManagerT) etcdDiscovery(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r,
		"https://github.com/giantswarm/mayu/blob/master/docs/etcd-discovery.md",
		http.StatusMovedPermanently,
	)
}

func (mgr *pxeManagerT) etcdDiscoveryNewCluster(w http.ResponseWriter, r *http.Request) {
	var err error
	size := 3
	s := r.FormValue("size")
	if s != "" {
		size, err = strconv.Atoi(s)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	token, err := mgr.cluster.GenerateEtcdCluster(size)

	if err != nil {
		glog.V(2).Infof("Unable to generate token '%v'", err)
		http.Error(w, "Unable to generate token", 400)
		return
	}

	err = mgr.cluster.Commit(fmt.Sprintf("add new etcd cluster '%s'", token))
	if err != nil {
		glog.V(2).Infof("Unable to store new etcd cluster information '%v'", err)
		http.Error(w, "Unable to store new etcd cluster information", 400)
		return
	}
	glog.V(2).Infof("New cluster created '%s'", token)

	fmt.Fprintf(w, "%s/%s", mgr.etcdDiscoveryBaseURL(), token)
}

func (mgr *pxeManagerT) etcdDiscoveryBaseURL() string {
	return fmt.Sprintf("%s/etcd", mgr.thisHost())
}

func (mgr *pxeManagerT) etcdDiscoveryToken(w http.ResponseWriter, r *http.Request) {
	token := mux.Vars(r)["token"]

	cluster, ok := mgr.cluster.Config.EtcdClusters[token]

	if !ok {
		mgr.etcdDiscoveryClusterNotFound(w, token)
		return
	}

	clusterKey := fmt.Sprintf("/_etcd/registry/%s", token)

	resp := EtcdResponse{
		Action: "get",
		Node: &EtcdNode{
			Key:   clusterKey,
			Dir:   true,
			Nodes: []*EtcdNode{},
		},
	}

	for k, v := range cluster.Machines {
		resp.Node.Nodes = append(resp.Node.Nodes, &EtcdNode{
			Key:   fmt.Sprintf("%s/%s", clusterKey, k),
			Value: fmt.Sprintf("%s=%s", v.Name, v.PeerURL),
		})
	}

	marshalled, err := json.Marshal(resp)
	if err != nil {
		glog.V(2).Infof("Can't encode response '%v'", err)
		http.Error(w, "Server error", 500)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	fmt.Fprint(w, string(marshalled))
}

func (mgr *pxeManagerT) etcdDiscoveryAddMachine(w http.ResponseWriter, r *http.Request) {
	token := mux.Vars(r)["token"]

	cluster, ok := mgr.cluster.Config.EtcdClusters[token]

	if !ok {
		mgr.etcdDiscoveryClusterNotFound(w, token)
		return
	}
	//var prevIgnore, prevExist, prevNoExist bool
	var prevNoExist bool

	machine := mux.Vars(r)["machine"]
	p := r.FormValue("prevExist")

	switch p {
	//case "":
	//	prevIgnore = true
	//case "true":
	//	prevExist = true
	case "false":
		prevNoExist = true
	default:
		http.Error(w, fmt.Sprintf("Bad parameter: prevExists should be boolean. '%s' given.", p), http.StatusBadRequest)
		return
	}

	if _, ok := cluster.Machines[machine]; ok && prevNoExist {
		mgr.etcdDiscoveryMachineExists(w, token, machine)
		return
	}

	// {"action":"create","node":{"key":"/_etcd/registry/be7acd8eccb86a1e7edc6d702fa51517/foobar","value":"FOO=http://1.2.3.4:4001","modifiedIndex":1124853439,"createdIndex":1124853439}}
	// {"errorCode":105,"message":"Key already exists","cause":"/_etcd/registry/be7acd8eccb86a1e7edc6d702fa51517/foobar","index":1124860329}

	value := r.FormValue("value")
	members := strings.Split(value, "&")

	// only accept the first member definition in the value (discovery.etcd.io does the same)
	member := members[0]
	m := strings.Split(member, "=")
	if len(m) < 2 {
		http.Error(w, fmt.Sprintf("Bad parameter: member value needs to have format 'name=peer_url'. '%s' given.", member), http.StatusBadRequest)
		return
	}
	cluster.Machines[machine] = &hostmgr.EtcdMachine{
		Name:    m[0],
		PeerURL: m[1],
	}

	err := mgr.cluster.Commit(fmt.Sprintf("add machine '%s' to etcd cluster '%s'", machine, token))
	if err != nil {
		glog.V(2).Infof("Unable to store etcd cluster information '%v'", err)
		http.Error(w, "Unable to store etcd cluster information", 400)
		return
	}

	resp := EtcdResponse{
		Action: "create",
		Node: &EtcdNode{
			Key:   fmt.Sprintf("/_etcd/registry/%s/%s", token, machine),
			Value: member,
		},
	}

	marshalled, err := json.Marshal(resp)
	if err != nil {
		glog.V(2).Infof("Can't encode response '%v'", err)
		http.Error(w, "Server error", 500)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	fmt.Fprint(w, string(marshalled))
}

func (mgr *pxeManagerT) etcdDiscoveryMachine(w http.ResponseWriter, r *http.Request) {
	token := mux.Vars(r)["token"]
	machine := mux.Vars(r)["machine"]

	cluster, ok := mgr.cluster.Config.EtcdClusters[token]
	if !ok {
		mgr.etcdDiscoveryClusterNotFound(w, token)
		return
	}

	var clusterMachine *hostmgr.EtcdMachine
	clusterMachine, ok = cluster.Machines[machine]
	if !ok {
		mgr.etcdDiscoveryMachineNotFound(w, token, machine)
		return
	}

	resp := EtcdResponse{
		Action: "get",
		Node: &EtcdNode{
			Key:   fmt.Sprintf("/_etcd/registry/%s/%s", token, machine),
			Value: fmt.Sprintf("%s=%s", clusterMachine.Name, clusterMachine.PeerURL),
		},
	}

	marshalled, err := json.Marshal(resp)
	if err != nil {
		glog.V(2).Infof("Can't encode response '%v'", err)
		http.Error(w, "Server error", 500)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	fmt.Fprint(w, string(marshalled))
}

func (mgr *pxeManagerT) etcdDiscoveryDeleteMachine(w http.ResponseWriter, r *http.Request) {
}

func (mgr *pxeManagerT) etcdDiscoveryTokenSize(w http.ResponseWriter, r *http.Request) {
	token := mux.Vars(r)["token"]

	cluster, ok := mgr.cluster.Config.EtcdClusters[token]

	if !ok {
		mgr.etcdDiscoveryClusterNotFound(w, token)
		return
	}

	resp := EtcdResponse{
		Action: "get",
		Node: &EtcdNode{
			Key:   fmt.Sprintf("/_etcd/registry/%s/_config/size", token),
			Value: strconv.Itoa(cluster.Size),
		},
	}

	marshalled, err := json.Marshal(resp)
	if err != nil {
		glog.V(2).Infof("Can't encode response '%v'", err)
		http.Error(w, "Server error", 500)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	fmt.Fprint(w, string(marshalled))
}

func (mgr *pxeManagerT) etcdDiscoveryUpdateTokenSize(w http.ResponseWriter, r *http.Request) {
	token := mux.Vars(r)["token"]

	cluster, ok := mgr.cluster.Config.EtcdClusters[token]

	if !ok {
		mgr.etcdDiscoveryClusterNotFound(w, token)
		return
	}

	s := r.FormValue("value")
	if s == "" {
		http.Error(w, "No size given", 400)
		return
	}
	size, err := strconv.Atoi(s)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	cluster.Size = size

	err = mgr.cluster.Commit(fmt.Sprintf("update etcd cluster size '%s' to '%d'", token, size))
	if err != nil {
		glog.V(2).Infof("Unable to store etcd cluster information '%v'", err)
		http.Error(w, "Unable to store etcd cluster information", 400)
		return
	}

	resp := EtcdResponse{
		Action: "get",
		Node: &EtcdNode{
			Key:   fmt.Sprintf("/_etcd/registry/%s/_config/size", token),
			Value: strconv.Itoa(cluster.Size),
		},
	}

	marshalled, err := json.Marshal(resp)
	if err != nil {
		glog.V(2).Infof("Can't encode response '%v'", err)
		http.Error(w, "Server error", 500)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	fmt.Fprint(w, string(marshalled))
}

func (mgr *pxeManagerT) etcdDiscoveryClusterNotFound(w http.ResponseWriter, token string) {
	resp := EtcdResponseError{
		ErrorCode: 100,
		Message:   "Key not found",
		Cause:     fmt.Sprintf("/_etcd/registry/%s", token),
	}

	marshalled, err := json.Marshal(resp)
	if err != nil {
		glog.V(2).Infof("Can't encode response '%v'", err)
		http.Error(w, "Server error", 500)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	glog.V(2).Infof("Cluster not found '%s'", token)
	http.Error(w, string(marshalled), 404)
}

func (mgr *pxeManagerT) etcdDiscoveryMachineNotFound(w http.ResponseWriter, token, machine string) {
	resp := EtcdResponseError{
		ErrorCode: 100,
		Message:   "Key not found",
		Cause:     fmt.Sprintf("/_etcd/registry/%s/%s", token, machine),
	}

	marshalled, err := json.Marshal(resp)
	if err != nil {
		glog.V(2).Infof("Can't encode response '%v'", err)
		http.Error(w, "Server error", 500)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	glog.V(2).Infof("Machine not found '%s/%s'", token, machine)
	http.Error(w, string(marshalled), 404)
}

func (mgr *pxeManagerT) etcdDiscoveryMachineExists(w http.ResponseWriter, token, machine string) {
	resp := EtcdResponseError{
		ErrorCode: 105,
		Message:   "Key already exists",
		Cause:     fmt.Sprintf("/_etcd/registry/%s/%s", token, machine),
	}

	marshalled, err := json.Marshal(resp)
	if err != nil {
		glog.V(2).Infof("Can't encode response '%v'", err)
		http.Error(w, "Server error", 500)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	glog.V(2).Infof("Machine '%s' already exists in etcd cluster '%s'", machine, token)
	http.Error(w, string(marshalled), 412)
}
