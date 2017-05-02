package client_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strconv"
	"testing"

	"github.com/giantswarm/mayu/client"
	"github.com/giantswarm/mayu/hostmgr"
)

type testResponse struct {
	Body   []byte
	Header http.Header
	Method string
	Path   string
}

func urlToHostPort(t *testing.T, URL string) (string, string) {
	u, err := url.Parse(URL)
	if err != nil {
		t.Fatalf("url.Parse returned error: %#v", err)
	}
	host, port, err := net.SplitHostPort(u.Host)
	if err != nil {
		t.Fatalf("net.SplitHostPort returned error: %#v", err)
	}

	return host, port
}

func newClientAndServer(t *testing.T, handler http.Handler) (*client.Client, *httptest.Server) {
	ts := httptest.NewServer(handler)

	host, port := urlToHostPort(t, ts.URL)
	ui, err := strconv.ParseUint(port, 10, 16)
	if err != nil {
		t.Fatalf("strconv.ParseUint returned error: %#v", err)
	}

	client, err := client.New("http", host, uint16(ui))
	if err != nil {
		t.Fatalf("client.New returned error: %#v", err)
	}

	return client, ts
}

//
// Client.SetMetadata
//

// Test_Client_001 checks for Client.SetMetadata to provide proper information
// to the server as expected.
func Test_Client_001(t *testing.T) {
	var response testResponse

	newClient, ts := newClientAndServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		r.Body.Close()
		if err != nil {
			t.Fatal(err)
		}

		response = testResponse{
			Body:   body,
			Header: r.Header,
			Method: r.Method,
			Path:   r.URL.Path,
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer ts.Close()

	err := newClient.SetMetadata("serial", "key1=value1,key2=value2")
	if err != nil {
		t.Fatalf("Client.SetMetadata returned error: %#v", err)
	}

	data, err := json.Marshal(hostmgr.Host{
		FleetMetadata: []string{"key1=value1", "key2=value2"},
	})
	if err != nil {
		t.Fatalf("json.Marshal returned error: %#v", err)
	}
	if string(response.Body) != string(data) {
		t.Fatalf("expected response body to be '%s', got '%s'", string(response.Body), string(data))
	}

	assertHeader(t, response, "content-type", []string{"application/json"})
	assertMethod(t, response, "PUT")
	assertPath(t, response, fmt.Sprintf("/admin/host/%s/set_metadata", "serial"))
}

// Test_Client_002 checks for Client.SetMetadata to provide proper error
// information to the client as expected, when there are errors returned from
// the server.
func Test_Client_002(t *testing.T) {
	newClient, ts := newClientAndServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
	}))
	defer ts.Close()

	err := newClient.SetMetadata("serial", "key1=value1,key2=value2")
	if err == nil {
		t.Fatalf("Client.SetMetadata NOT returned error")
	}
}

// Test_Client_003 checks for Client.SetMetadata to provide proper error
// information to the client as expected, when there is no server running.
func Test_Client_003(t *testing.T) {
	newClient, ts := newClientAndServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
	}))
	// Immediatelly close the server.
	ts.Close()

	err := newClient.SetMetadata("serial", "key1=value1,key2=value2")
	if err == nil {
		t.Fatalf("Client.SetMetadata NOT returned error")
	}
}

//
// Client.SetProviderId
//

// Test_Client_004 checks for Client.SetProviderId to provide proper information
// to the server as expected.
func Test_Client_004(t *testing.T) {
	var response testResponse

	newClient, ts := newClientAndServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		r.Body.Close()
		if err != nil {
			t.Fatal(err)
		}

		response = testResponse{
			Body:   body,
			Header: r.Header,
			Method: r.Method,
			Path:   r.URL.Path,
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer ts.Close()

	err := newClient.SetProviderId("serial", "provider-id")
	if err != nil {
		t.Fatalf("Client.SetProviderId returned error: %#v", err)
	}

	data, err := json.Marshal(hostmgr.Host{
		ProviderId: "provider-id",
	})
	if err != nil {
		t.Fatalf("json.Marshal returned error: %#v", err)
	}
	if string(response.Body) != string(data) {
		t.Fatalf("expected response body to be '%s', got '%s'", string(response.Body), string(data))
	}

	assertHeader(t, response, "content-type", []string{"application/json"})
	assertMethod(t, response, "PUT")
	assertPath(t, response, fmt.Sprintf("/admin/host/%s/set_provider_id", "serial"))
}

// Test_Client_005 checks for Client.SetProviderId to provide proper error
// information to the client as expected, when there are errors returned from
// the server.
func Test_Client_005(t *testing.T) {
	newClient, ts := newClientAndServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
	}))
	defer ts.Close()

	err := newClient.SetProviderId("serial", "provider-id")
	if err == nil {
		t.Fatalf("Client.SetProviderId NOT returned error")
	}
}

// Test_Client_006 checks for Client.SetProviderId to provide proper error
// information to the client as expected, when there is no server running.
func Test_Client_006(t *testing.T) {
	newClient, ts := newClientAndServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
	}))
	// Immediatelly close the server.
	ts.Close()

	err := newClient.SetProviderId("serial", "provider-id")
	if err == nil {
		t.Fatalf("Client.SetProviderId NOT returned error")
	}
}

//
// Client.SetIPMIAddr
//

// Test_Client_007 checks for Client.SetIPMIAddr to provide proper information
// to the server as expected.
func Test_Client_007(t *testing.T) {
	var response testResponse

	newClient, ts := newClientAndServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		r.Body.Close()
		if err != nil {
			t.Fatal(err)
		}

		response = testResponse{
			Body:   body,
			Header: r.Header,
			Method: r.Method,
			Path:   r.URL.Path,
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer ts.Close()

	err := newClient.SetIPMIAddr("serial", "127.0.0.1")
	if err != nil {
		t.Fatalf("Client.SetIPMIAddr returned error: %#v", err)
	}

	data, err := json.Marshal(hostmgr.Host{
		IPMIAddr: net.ParseIP("127.0.0.1"),
	})
	if err != nil {
		t.Fatalf("json.Marshal returned error: %#v", err)
	}
	if string(response.Body) != string(data) {
		t.Fatalf("expected response body to be '%s', got '%s'", string(response.Body), string(data))
	}

	assertHeader(t, response, "content-type", []string{"application/json"})
	assertMethod(t, response, "PUT")
	assertPath(t, response, fmt.Sprintf("/admin/host/%s/set_ipmi_addr", "serial"))
}

// Test_Client_008 checks for Client.SetIPMIAddr to provide proper error
// information to the client as expected, when there are errors returned from
// the server.
func Test_Client_008(t *testing.T) {
	newClient, ts := newClientAndServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
	}))
	defer ts.Close()

	err := newClient.SetIPMIAddr("serial", "127.0.0.1")
	if err == nil {
		t.Fatalf("Client.SetIPMIAddr NOT returned error")
	}
}

// Test_Client_009 checks for Client.SetIPMIAddr to provide proper error
// information to the client as expected, when there is no server running.
func Test_Client_009(t *testing.T) {
	newClient, ts := newClientAndServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
	}))
	// Immediatelly close the server.
	ts.Close()

	err := newClient.SetIPMIAddr("serial", "127.0.0.1")
	if err == nil {
		t.Fatalf("Client.SetIPMIAddr NOT returned error")
	}
}

//
// Client.SetCabinet
//

// Test_Client_010 checks for Client.SetCabinet to provide proper information
// to the server as expected.
func Test_Client_010(t *testing.T) {
	var response testResponse

	newClient, ts := newClientAndServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		r.Body.Close()
		if err != nil {
			t.Fatal(err)
		}

		response = testResponse{
			Body:   body,
			Header: r.Header,
			Method: r.Method,
			Path:   r.URL.Path,
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer ts.Close()

	err := newClient.SetCabinet("serial", "101")
	if err != nil {
		t.Fatalf("Client.SetCabinet returned error: %#v", err)
	}

	data, err := json.Marshal(hostmgr.Host{
		Cabinet: uint(101),
	})
	if err != nil {
		t.Fatalf("json.Marshal returned error: %#v", err)
	}
	if string(response.Body) != string(data) {
		t.Fatalf("expected response body to be '%s', got '%s'", string(response.Body), string(data))
	}

	assertHeader(t, response, "content-type", []string{"application/json"})
	assertMethod(t, response, "PUT")
	assertPath(t, response, fmt.Sprintf("/admin/host/%s/set_cabinet", "serial"))
}

// Test_Client_011 checks for Client.SetCabinet to provide proper error
// information to the client as expected, when there are errors returned from
// the server.
func Test_Client_011(t *testing.T) {
	newClient, ts := newClientAndServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
	}))
	defer ts.Close()

	err := newClient.SetCabinet("serial", "101")
	if err == nil {
		t.Fatalf("Client.SetCabinet NOT returned error")
	}
}

// Test_Client_012 checks for Client.SetCabinet to provide proper error
// information to the client as expected, when there is no server running.
func Test_Client_012(t *testing.T) {
	newClient, ts := newClientAndServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
	}))
	// Immediatelly close the server.
	ts.Close()

	err := newClient.SetCabinet("serial", "101")
	if err == nil {
		t.Fatalf("Client.SetCabinet NOT returned error")
	}
}

//
// Client.List
//

// Test_Client_013 checks for Client.List to provide proper information
// to the server as expected.
func Test_Client_013(t *testing.T) {
	var response testResponse
	expectedList := []hostmgr.Host{
		hostmgr.Host{
			Id:   101,
			Name: "test-host-101",
		},
		hostmgr.Host{
			Id:   102,
			Name: "test-host-102",
		},
		hostmgr.Host{
			Id:   103,
			Name: "test-host-103",
		},
	}

	newClient, ts := newClientAndServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response = testResponse{
			Method: r.Method,
			Path:   r.URL.Path,
		}

		if err := json.NewEncoder(w).Encode(expectedList); err != nil {
			t.Fatalf("json.NewEncoder(w).Encode returned error: %#v", err)
		}
	}))
	defer ts.Close()

	list, err := newClient.List()
	if err != nil {
		t.Fatalf("Client.List returned error: %#v", err)
	}

	if !reflect.DeepEqual(list, expectedList) {
		t.Fatalf("expected %#v got %#v", expectedList, list)
	}

	assertMethod(t, response, "GET")
	assertPath(t, response, "/admin/hosts")
}

// Test_Client_014 checks for Client.List to provide proper error
// information to the client as expected, when there are errors returned from
// the server.
func Test_Client_014(t *testing.T) {
	newClient, ts := newClientAndServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
	}))
	defer ts.Close()

	_, err := newClient.List()
	if err == nil {
		t.Fatalf("Client.List NOT returned error")
	}
}

// Test_Client_015 checks for Client.List to provide proper error
// information to the client as expected, when there is no server running.
func Test_Client_015(t *testing.T) {
	newClient, ts := newClientAndServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
	}))
	// Immediatelly close the server.
	ts.Close()

	_, err := newClient.List()
	if err == nil {
		t.Fatalf("Client.List NOT returned error")
	}
}

//
// Client.Status
//

// Test_Client_016 checks for Client.Status to provide proper information
// to the server as expected.
func Test_Client_016(t *testing.T) {
	var response testResponse
	returnedList := []hostmgr.Host{
		hostmgr.Host{
			Id:     101,
			Serial: "serial-101",
			Name:   "test-host-101",
		},
		hostmgr.Host{
			Id:     102,
			Serial: "serial-102",
			Name:   "test-host-102",
		},
		hostmgr.Host{
			Id:     103,
			Serial: "serial-103",
			Name:   "test-host-103",
		},
	}

	expectedHost := hostmgr.Host{
		Id:     102,
		Serial: "serial-102",
		Name:   "test-host-102",
	}

	newClient, ts := newClientAndServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response = testResponse{
			Method: r.Method,
			Path:   r.URL.Path,
		}

		if err := json.NewEncoder(w).Encode(returnedList); err != nil {
			t.Fatalf("json.NewEncoder(w).Encode returned error: %#v", err)
		}
	}))
	defer ts.Close()

	host, err := newClient.Status("serial-102")
	if err != nil {
		t.Fatalf("Client.Status returned error: %#v", err)
	}

	if !reflect.DeepEqual(host, expectedHost) {
		t.Fatalf("expected %#v got %#v", expectedHost, host)
	}

	assertMethod(t, response, "GET")
	assertPath(t, response, "/admin/hosts")
}

// Test_Client_017 checks for Client.Status to provide proper error
// information to the client as expected, when there are errors returned from
// the server.
func Test_Client_017(t *testing.T) {
	newClient, ts := newClientAndServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
	}))
	defer ts.Close()

	_, err := newClient.Status("serial")
	if err == nil {
		t.Fatalf("Client.Status NOT returned error")
	}
}

// Test_Client_018 checks for Client.Status to provide proper error
// information to the client as expected, when there is no server running.
func Test_Client_018(t *testing.T) {
	newClient, ts := newClientAndServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
	}))
	// Immediatelly close the server.
	ts.Close()

	_, err := newClient.Status("serial")
	if err == nil {
		t.Fatalf("Client.Status NOT returned error")
	}
}
