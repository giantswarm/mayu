package hostmgr

import "fmt"

type hostState int

//XXX TODO intermediate states:
//      => initially `installing` would be aok
//        => eventually a `decommissioning` would be nice to have

const (
	Unknown hostState = iota
	Configured
	Installing
	Installed
	Running
)

func HostStateMap() map[hostState]string {
	return map[hostState]string{
		Unknown:    `"unknown"`,
		Configured: `"configured"`,
		Installing: `"installing"`,
		Installed:  `"installed"`,
		Running:    `"running"`,
	}
}

func (s hostState) MarshalJSON() ([]byte, error) {
	m := HostStateMap()
	if stringVal, ok := m[s]; ok {
		return []byte(stringVal), nil
	}

	return []byte{}, fmt.Errorf("don't know how to marshal '%d'", s)
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
	case `"installed"`:
		*s = Installed
	case `"running"`:
		*s = Running
	default:
		return fmt.Errorf("don't know how to unmarshal '%+v'", b)
	}
	return nil
}
