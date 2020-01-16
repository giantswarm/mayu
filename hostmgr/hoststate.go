package hostmgr

import (
	"fmt"
	"github.com/giantswarm/microerror"
)

type hostState int

const (
	Unknown hostState = iota
	Configured
	Installing
	Running
)

func HostStateMap() map[hostState]string {
	return map[hostState]string{
		Unknown:    `"unknown"`,
		Configured: `"configured"`,
		Installing: `"installing"`,
		Running:    `"running"`,
	}
}

func (s hostState) MarshalJSON() ([]byte, error) {
	m := HostStateMap()
	if stringVal, ok := m[s]; ok {
		return []byte(stringVal), nil
	}

	return []byte{}, microerror.Mask(fmt.Errorf("don't know how to marshal '%d'", s))
}

func HostState(state string) (hostState, error) {
	switch state {
	case "unknown":
		return Unknown, nil
	case "configured":
		return Configured, nil
	case "installing":
		return Installing, nil
	case "running":
		return Running, nil
	default:
		return -1, microerror.Mask(fmt.Errorf("wrong host state '%s'", state))
	}
}

func (s *hostState) UnmarshalJSON(b []byte) error {
	str := string(b)
	switch str {
	case `"unknown"`:
		*s = Unknown
	case `"configured"`:
		*s = Configured
	case `"installing"`:
		*s = Installing
	case `"running"`:
		*s = Running
	default:
		return microerror.Mask(fmt.Errorf("don't know how to unmarshal '%+v'", b))
	}
	return nil
}
