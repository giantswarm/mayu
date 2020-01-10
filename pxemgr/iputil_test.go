package pxemgr

import (
	"testing"

	"net"
)

func TestIPUtil_IncIP(t *testing.T) {
	// Create test IPs. Each slice represents a test case. The first IP is the
	// input. The last IP is the expected output.
	IPs := [][]net.IP{
		[]net.IP{net.ParseIP("0.0.0.0"), net.ParseIP("0.0.0.1")},
		[]net.IP{net.ParseIP("0.0.0.1"), net.ParseIP("0.0.0.2")},
		[]net.IP{net.ParseIP("0.0.1.0"), net.ParseIP("0.0.1.1")},
		[]net.IP{net.ParseIP("3.0.0.0"), net.ParseIP("3.0.0.1")},
		[]net.IP{net.ParseIP("0.0.0.255"), net.ParseIP("0.0.1.0")},
		[]net.IP{net.ParseIP("0.0.255.255"), net.ParseIP("0.1.0.0")},
		[]net.IP{net.ParseIP("0.255.255.255"), net.ParseIP("1.0.0.0")},
		[]net.IP{net.ParseIP("0.0.255.0"), net.ParseIP("0.0.255.1")},
		[]net.IP{net.ParseIP("0.255.0.0"), net.ParseIP("0.255.0.1")},
		[]net.IP{net.ParseIP("255.0.0.0"), net.ParseIP("255.0.0.1")},
	}

	for _, testCase := range IPs {
		input := testCase[0]
		expected := testCase[1]
		output := incIP(input)
		if output.String() != expected.String() {
			t.Fatalf("expected IP '%s' to be incremented to IP '%s', got IP '%s'", input, expected, output)
		}
	}
}

func TestIPUtil_IPLessThanOrEqual(t *testing.T) {
	// Create test IPs. Each slice represents a test case. The first IP is the
	// smaller one. The last IP is the bigger or equal one.
	IPs := [][]net.IP{
		[]net.IP{net.ParseIP("0.0.0.0"), net.ParseIP("0.0.0.1")},
		[]net.IP{net.ParseIP("0.0.0.0"), net.ParseIP("0.0.0.0")}, // equal
		[]net.IP{net.ParseIP("0.0.0.1"), net.ParseIP("0.0.0.2")},
		[]net.IP{net.ParseIP("0.0.1.0"), net.ParseIP("0.0.1.1")},
		[]net.IP{net.ParseIP("0.0.1.0"), net.ParseIP("0.0.1.0")}, // equal
		[]net.IP{net.ParseIP("3.0.0.0"), net.ParseIP("3.0.0.1")},
		[]net.IP{net.ParseIP("3.0.0.0"), net.ParseIP("3.0.0.0")}, // equal
		[]net.IP{net.ParseIP("0.0.0.255"), net.ParseIP("0.0.1.0")},
		[]net.IP{net.ParseIP("0.0.255.255"), net.ParseIP("0.1.0.0")},
		[]net.IP{net.ParseIP("0.255.255.255"), net.ParseIP("1.0.0.0")},
		[]net.IP{net.ParseIP("0.255.255.255"), net.ParseIP("0.255.255.255")}, // equal
		[]net.IP{net.ParseIP("0.0.255.0"), net.ParseIP("0.0.255.1")},
		[]net.IP{net.ParseIP("0.255.0.0"), net.ParseIP("0.255.0.1")},
		[]net.IP{net.ParseIP("255.0.0.0"), net.ParseIP("255.0.0.1")},
	}

	for _, testCase := range IPs {
		smaller := testCase[0]
		biggerOrEqual := testCase[1]
		if !ipLessThanOrEqual(smaller, biggerOrEqual) {
			t.Fatalf("expected IP '%s' to be less then, or equal to IP '%s', but it was detected to be greater", smaller, biggerOrEqual)
		}
	}
}

func TestIPUtil_IPMoreThanOrEqual(t *testing.T) {
	// Create test IPs. Each slice represents a test case. The first IP is the
	// bigger one. The last IP is the smaller or equal one.
	IPs := [][]net.IP{
		[]net.IP{net.ParseIP("0.0.0.1"), net.ParseIP("0.0.0.0")},
		[]net.IP{net.ParseIP("0.0.0.0"), net.ParseIP("0.0.0.0")}, // equal
		[]net.IP{net.ParseIP("0.0.0.2"), net.ParseIP("0.0.0.1")},
		[]net.IP{net.ParseIP("0.0.1.1"), net.ParseIP("0.0.1.0")},
		[]net.IP{net.ParseIP("0.0.1.0"), net.ParseIP("0.0.1.0")}, // equal
		[]net.IP{net.ParseIP("3.0.0.1"), net.ParseIP("3.0.0.0")},
		[]net.IP{net.ParseIP("3.0.0.0"), net.ParseIP("3.0.0.0")}, // equal
		[]net.IP{net.ParseIP("0.0.1.0"), net.ParseIP("0.0.0.255")},
		[]net.IP{net.ParseIP("0.1.0.0"), net.ParseIP("0.0.255.255")},
		[]net.IP{net.ParseIP("1.0.0.0"), net.ParseIP("0.255.255.255")},
		[]net.IP{net.ParseIP("0.255.255.255"), net.ParseIP("0.255.255.255")}, // equal
		[]net.IP{net.ParseIP("0.0.255.1"), net.ParseIP("0.0.255.0")},
		[]net.IP{net.ParseIP("0.255.0.1"), net.ParseIP("0.255.0.0")},
		[]net.IP{net.ParseIP("255.0.0.1"), net.ParseIP("255.0.0.0")},
	}

	for _, testCase := range IPs {
		bigger := testCase[0]
		smallerOrEqual := testCase[1]

		if !ipMoreThanOrEqual(bigger, smallerOrEqual) {
			t.Fatalf("expected IP '%s' to be bigger then, or equal to IP '%s', but it was detected to be smaller", bigger, smallerOrEqual)
		}
	}
}
