package whitelist

import (
	"crypto/sha1"
	"github.com/cloudendpoints/mixologist/mixologist"
	"github.com/golang/glog"
	sc "google/api/servicecontrol/v1"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// whitelist -- typed atomic accessor for whitelist
func (c *checker) whitelist() []*net.IPNet {
	return c.atomicWhitelist.Load().([]*net.IPNet)
}

// setWhitelist -- typed atomic setter for whitelist
func (c *checker) setWhitelist(wl []*net.IPNet) {
	c.atomicWhitelist.Store(wl)
}

func (c *checker) checkWhiteList(ip string) bool {
	ipa := net.ParseIP(ip)
	for _, ipnet := range c.whitelist() {
		if ipnet.Contains(ipa) {
			return true
		}
	}
	return false
}

// Check -- Check if client ip is on the whitelist
func (c *checker) Check(cr *sc.CheckRequest) (*sc.CheckError, error) {
	//Check: service_name:"owner-1470410002014.appspot.com" operation:<operation_id:"c37d4302-66bd-4f34-8ed7-07b36d941fcd" operation_name:"ListShelves" consumer_id:"project:mixologist-142215" start_time:<seconds:1475272937 nanos:398591000 > end_time:<seconds:1475273136 nanos:719613000 > labels:<key:"servicecontrol.googleapis.com/caller_ip" value:"10.128.0.2" > labels:<key:"servicecontrol.googleapis.com/service_agent" value:"ESP/0.3.7" > labels:<key:"servicecontrol.googleapis.com/user_agent" value:"ESP" > >
	if ip, found := cr.GetOperation().GetLabels()[ClientIPKey]; found {
		if c.checkWhiteList(ip) {
			return nil, nil
		}
		glog.V(1).Infof(ip, " Not in whitelist ", c.whitelist)
		return IPBlockedCheckError, nil
	}

	return nil, ErrClientIPMissing
}

// updateConfig -- fetch list from backend and populate datastructure
func (c *checker) updateConfigLoop() {
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()

	clnt := &http.Client{
		Timeout: time.Second * 5,
	}
	// nearly synchronous config fetch
	c.updateConfig(clnt)
	done := false

	for !done {
		select {
		case <-ticker.C:
			c.updateConfig(clnt)
		case <-c.closing:
			done = true
		}
	}
	glog.V(2).Info("Unloaded")
}

// updateConfig -- fetch list from backend and populate datastructure
func (c *checker) updateConfig(clnt *http.Client) error {
	resp, err := clnt.Get(c.backend)
	if err != nil {
		glog.Warning("Could not connect to ", c.backend, " ", err)
		return err
	}
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		glog.Warning("Could not read from ", c.backend, " ", err)
		return err
	}

	newsha := sha1.Sum(buf)
	// c.fetchedSha is only read and written by this function
	// in a single thread
	if newsha != c.fetchedSha {
		glog.Infoln("Fetched new config from ", c.backend)
		wlcfg := CfgList{}
		err = yaml.Unmarshal(buf, &wlcfg)
		if err != nil || len(wlcfg.WhiteList) == 0 {
			glog.Warning("Could not unmarshal ", c.backend, " ", err)
			return err
		}
		// Now create a new map and install it
		c.setWhitelist(buildWhiteList(wlcfg.WhiteList...))
		c.fetchedSha = newsha
	}
	return nil
}

// buildWhiteList -- convert a list of strings to ipnets
func buildWhiteList(whitelist ...string) []*net.IPNet {
	wl := make([]*net.IPNet, 0, len(whitelist))
	for _, ip := range whitelist {
		if !strings.Contains(ip, "/") {
			ip += "/32"
		}
		_, ipnet, err := net.ParseCIDR(ip)
		if err != nil {
			glog.Warningf("Unable to parse %s -- %v", ip, err)
			continue
		}
		wl = append(wl, ipnet)
	}
	glog.V(1).Info("New whitelist", wl)
	return wl
}

func (c *checker) Name() string {
	return Name
}

func (c *checker) Unload() {
	close(c.closing)
}

func init() {
	mixologist.RegisterChecker(Name, new(builder))
}

// BuildChecker -- exported method
func (b *builder) BuildChecker(cfg interface{}) (mixologist.Checker, error) {
	wlcfg := cfg.(*Config)
	chk := &checker{
		backend: wlcfg.ProviderURL,
		closing: make(chan bool),
	}
	// install an empty list
	chk.atomicWhitelist.Store([]*net.IPNet{})
	go chk.updateConfigLoop()
	return chk, nil
}

// ConfigStruct -- return pointer to Config struct
func (b *builder) ConfigStruct() interface{} {
	return &Config{}
}

// ValidateConfig -- validate given config
func (b *builder) ValidateConfig(cfg interface{}) error {
	wlcfg := cfg.(*Config)
	_, err := url.Parse(wlcfg.ProviderURL)
	return err
}
