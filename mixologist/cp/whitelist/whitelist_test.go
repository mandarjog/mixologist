package whitelist

import (
	g "github.com/onsi/gomega"
	sc "google/api/servicecontrol/v1"
	"gopkg.in/yaml.v2"
	"net/http"
	"net/http/httptest"
	"runtime"
	"sync/atomic"
	"testing"
)

func build(url string) (*checker, error) {
	chk, err := new(builder).BuildChecker(&Config{
		ProviderURL: url,
	})

	return chk.(*checker), err
}

func TestWhitelistFetch(t *testing.T) {
	cfg := CfgList{
		WhiteList: []string{"10.10.11.2", "10.10.11.3"},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var out []byte
		var err error
		if out, err = yaml.Marshal(cfg); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(out)
	}))
	defer ts.Close()
	wl := &checker{
		backend: ts.URL,
	}
	clnt := &http.Client{}
	err := wl.updateConfig(clnt)
	if err != nil {
		t.Errorf("Expected success, got %s", err)
	}

	for _, ip := range cfg.WhiteList {
		if !wl.checkWhiteList(ip) {
			t.Errorf("Failed: Expected %s in whitelist (%v)", ip, wl.whitelist())
		}
	}

	// change value on the server
	IPAddr := "202.54.10.2"

	cfg.WhiteList[0] = IPAddr
	err = wl.updateConfig(clnt)
	if err != nil {
		t.Errorf("Expected success, got %s", err)
	}
	if !wl.checkWhiteList(IPAddr) {
		t.Errorf("Failed: Expected %s in whitelist (%v)", IPAddr, wl.whitelist())
	}
}

func TestWhiteListUnload(t *testing.T) {
	g.RegisterTestingT(t)
	wl, err := build("")
	g.Expect(err).To(g.BeNil())
	var finalized atomic.Value
	finalized.Store(false)
	// check adapter is eventually unloaded
	// This test ensures that after unload is called
	// the adapter is removed from memory
	runtime.SetFinalizer(wl, func(ff interface{}) {
		finalized.Store(true)
	})
	wl.Unload()
	runtime.GC()
	g.Eventually(func() bool {
		runtime.GC()
		return finalized.Load().(bool)
	}).Should(g.BeTrue(), "Adapter was not fully unloaded")
}

func checkRequest(ipaddr string) *sc.CheckRequest {
	cr := &sc.CheckRequest{
		ServiceName: "testservice",
		Operation: &sc.Operation{
			Labels: make(map[string]string),
		},
	}

	if len(ipaddr) > 0 {
		cr.Operation.Labels[ClientIPKey] = ipaddr
	}

	return cr
}

func buildChecker(ipaddr ...string) *checker {
	wl := &checker{}
	wl.setWhitelist(buildWhiteList(ipaddr...))
	return wl
}

func testcase(checkerAddrs []string, addr string, expectedErr error, expectedCheckErr *sc.CheckError, msg string) {
	wl := buildChecker(checkerAddrs...)
	cr := checkRequest(addr)
	ce, err := wl.Check(cr)
	if expectedErr == nil {
		g.Expect(err).To(g.BeNil())
	} else {
		g.Expect(err).To(g.Equal(expectedErr))
	}
	g.Expect(ce).To(g.Equal(expectedCheckErr), msg)
}

func TestWhiteList(t *testing.T) {
	g.RegisterTestingT(t)

	IPAddr := "9.9.9.9"
	IPAddr1 := "9.9.9.1"

	testcase([]string{IPAddr}, IPAddr, nil, nil, IPAddr+" Should succeed")

	// send a blocked ip address
	testcase([]string{IPAddr}, IPAddr1, nil, IPBlockedCheckError, IPAddr1+" Should be blocked")

	// buildchecker to allow entire subnet
	testcase([]string{IPAddr + "/28"}, IPAddr1, nil, nil, IPAddr1+" Should succeed")
}

func TestWhiteListBadRequest(t *testing.T) {
	g.RegisterTestingT(t)
	badcr := checkRequest("")
	IPAddr := "9.9.9.9"
	wl := buildChecker(IPAddr)
	ce, err := wl.Check(badcr)
	g.Expect(err).To(g.Equal(ErrClientIPMissing))
	g.Expect(ce).To(g.BeNil(), IPAddr+" Should succeed")
}
