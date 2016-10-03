package whitelist

import (
	"crypto/sha1"
	"errors"
	sc "google/api/servicecontrol/v1"
	"net"
	"sync"
)

const (
	// Name -- name of this provider. TODO Should be namespaced.
	Name = "whitelist"
	// ClientIPKey - key used by service control to pass thru client ip
	ClientIPKey = "servicecontrol.googleapis.com/caller_ip"
	// IPBlockedErrorMsg -- error msg while rejecting
	IPBlockedErrorMsg = "IP address is not on the whitelist"
)

type (
	builder struct {
	}

	checker struct {
		backend    string
		whitelist  []*net.IPNet
		fetchedSha [sha1.Size]byte

		lock sync.RWMutex
	}

	// Config -- struct needed to configure this checker
	Config struct {
		ProviderURL string
	}
	// CfgList -- file format of the exteral file denoting a whitelist
	CfgList struct {
		WhiteList []string `yaml:"whitelist"`
	}
)

var (
	// IPBlockedCheckError -- prefined val for returning an error
	IPBlockedCheckError = &sc.CheckError{
		Code:   sc.CheckError_IP_ADDRESS_BLOCKED,
		Detail: IPBlockedErrorMsg,
	}
	// ErrClientIPMissing -- If client IP is missing, an error is thrown
	ErrClientIPMissing = errors.New(ClientIPKey + " Label not found")
)
