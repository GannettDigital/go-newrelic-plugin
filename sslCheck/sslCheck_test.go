package sslCheck

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/franela/goblin"
)

func TestValidateConfig(t *testing.T) {
	g := goblin.Goblin(t)
	g.Describe("ValidateConfig()", func() {
		expected := map[string]struct {
			ExpectedIsNil bool
			Config        Config
		}{
			"no":    {false, Config{Hosts: []string{}}},
			"Hosts": {true, Config{Hosts: []string{"example.com:443", "sub.domain.com:8443"}}},
		}
		for name, ex := range expected {
			desc := fmt.Sprintf("should return %v when %v fields are set", ex.ExpectedIsNil, name)
			g.It(desc, func() {
				valid := ValidateConfig(ex.Config)
				g.Assert(valid == nil).Equal(ex.ExpectedIsNil)
			})
		}
	})
}

func TestProcessHosts(t *testing.T) {
	g := goblin.Goblin(t)

	var tests = []struct {
		hosts              string
		expectedHosts      []string
		expectedErrorIsNil bool
	}{
		{
			hosts:              "example.com:443,sub.domain.com:8443,notValid",
			expectedHosts:      []string{"example.com:443", "sub.domain.com:8443"},
			expectedErrorIsNil: true,
		},
		{
			hosts:              "",
			expectedHosts:      []string{},
			expectedErrorIsNil: false,
		},
	}

	for _, test := range tests {
		g.Describe("ProcessHosts()", func() {
			g.It(fmt.Sprintf("When hosts are: %v, then expectedErrorIsNil should be: %v", test.hosts, test.expectedErrorIsNil), func() {
				result, valid := ProcessHosts(test.hosts)
				g.Assert(reflect.DeepEqual(valid, nil)).Equal(test.expectedErrorIsNil)
				g.Assert(reflect.DeepEqual(len(result), len(test.expectedHosts))).IsTrue()
			})
		})
	}
}

func TestValidHost(t *testing.T) {
	g := goblin.Goblin(t)

	var tests = []struct {
		host    string
		isValid bool
	}{
		{
			host:    "example.com:443",
			isValid: true,
		},
		{
			host:    "myhost",
			isValid: false,
		},
		{
			host:    "example.com",
			isValid: false,
		},
		{
			host:    "*****",
			isValid: false,
		},
		{
			host:    "sub.domain.com:8443",
			isValid: true,
		},
		{
			host:    "invalid.port.com:844322",
			isValid: true,
		},
	}

	for _, test := range tests {
		g.Describe("validHost()", func() {
			g.It(fmt.Sprintf("When host is: %v, then isValid should be: %v", test.host, test.isValid), func() {
				result := validHost(test.host)
				g.Assert(result == test.isValid).Equal(true)
			})
		})
	}
}
