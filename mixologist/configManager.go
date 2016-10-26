package mixologist

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"io/ioutil"
	"k8s.io/client-go/1.5/kubernetes"
	"k8s.io/client-go/1.5/tools/clientcmd"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/golang/glog"
	yaml "gopkg.in/yaml.v2"
)

const (
	ConfigMapScheme = "configmap"
)

type ConfigManager struct {
	cl         []ConfigChanger
	url        *url.URL
	clnt       *http.Client
	k8sClient  *kubernetes.Clientset
	fetchedSha [sha1.Size]byte
	closing    chan bool
}

func NewConfigManager(curl string, kubeconfig string) (*ConfigManager, error) {
	u, err := url.Parse(curl)
	if err != nil {
		return nil, err
	}

	cm := &ConfigManager{
		url: u,
		clnt: &http.Client{
			Timeout: time.Second * 5,
		},
		closing: make(chan bool),
	}

	if u.Scheme == ConfigMapScheme {
		config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, err
		}
		k8sClient, err := kubernetes.NewForConfig(config)
		if err != nil {
			return nil, err
		}
		cm.k8sClient = k8sClient
	}
	return cm, nil
}

func (c *ConfigManager) Register(cc ConfigChanger) {
	c.cl = append(c.cl, cc)
}

func (c *ConfigManager) Loop() {
	c.FetchAndNotify()
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()
	done := false

	for !done {
		select {
		case <-ticker.C:
			err := c.FetchAndNotify()
			if err != nil {
				glog.Warning(err)
			}
		case <-c.closing:
			done = true
		}
	}
}

func (c *ConfigManager) FetchAndNotify() error {
	var data []byte
	var err error
	var newsha [sha1.Size]byte

	if c.url.Scheme == ConfigMapScheme {
		mapname := strings.TrimLeft(c.url.Path, "/")
		cfg, err := c.k8sClient.ConfigMaps(c.url.Host).Get(mapname)
		if err != nil {
			return err
		}
		if len(cfg.Data) > 1 {
			glog.Warningf("map has multiple configs")
		}
		for _, v := range cfg.Data {
			data = []byte(v)
			break
		}
	} else {
		if strings.HasPrefix(c.url.Scheme, "http") {
			resp, err := c.clnt.Get(c.url.String())
			if err != nil {
				glog.Warning("Could not connect to ", c.url, " ", err)
				return err
			}
			if data, err = ioutil.ReadAll(resp.Body); err != nil {
				glog.Warning("Could not read from ", c.url, " ", err)
				return err
			}
			if resp.StatusCode != http.StatusOK {
				msg := fmt.Sprintf("Could not get config %s %v", c.url, resp.Status)
				glog.Warning(msg)
				return errors.New(msg)
			}
		} else {
			if data, err = ioutil.ReadFile(c.url.String()); err != nil {
				glog.Errorf("Unable to read %s:  %s", c.url, err)
				return err
			}
		}
	}
	newsha = sha1.Sum(data)
	// check if sha has changed
	if newsha == c.fetchedSha {
		glog.V(3).Infof("No change in config")
		return nil
	}
	osc := ServicesConfig{}
	yaml.Unmarshal(data, &osc)
	// have osc parsed
	ssc, erra := ConvertParams(osc, CheckerRegistry)
	if len(erra) > 0 {
		glog.Warningf("Unable to process some adapters, %s", erra)
	}
	glog.Infof("Installing new config from %s sha=%x ", c.url.String(), newsha)
	// notify
	c.fetchedSha = newsha
	for _, cc := range c.cl {
		cc.ConfigChange(&ssc)
	}
	return nil
}
