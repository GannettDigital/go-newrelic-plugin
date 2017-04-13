package sslCheck

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"reflect"
	"testing"
	"time"

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

func TestValidateCertificate(t *testing.T) {
	g := goblin.Goblin(t)

	testTimeNow := time.Now()
	var tests = []struct {
		description    string
		timeNow        time.Time
		cert           x509.Certificate
		isCertValid    bool
		expetedCertErr certError
	}{
		{
			description: "expires in 180 days",
			timeNow:     testTimeNow,
			cert: x509.Certificate{
				NotAfter: testTimeNow.AddDate(0, 0, 180),
				Subject: pkix.Name{
					CommonName: "ExpireIn180DaysCommonName",
				},
			},
			isCertValid: true,
			expetedCertErr: certError{
				CommonName:                 "ExpireIn180DaysCommonName",
				ExpirationDate:             testTimeNow.AddDate(0, 0, 180).String(),
				ExpiresInFiveDaysOrLess:    false,
				ExpiresInFifteenDaysOrLess: false,
				ExpiresInThirtyDaysOrLess:  false,
				ExpiresInSixtyDaysOrLess:   false,
			},
		},
		{
			description: "expires in 59 days",
			timeNow:     testTimeNow,
			cert: x509.Certificate{
				NotAfter: testTimeNow.AddDate(0, 0, 59),
				Subject: pkix.Name{
					CommonName: "ExpireIn59DaysCommonName",
				},
			},
			isCertValid: false,
			expetedCertErr: certError{
				CommonName:                 "ExpireIn59DaysCommonName",
				ExpirationDate:             testTimeNow.AddDate(0, 0, 59).String(),
				ExpiresInFiveDaysOrLess:    false,
				ExpiresInFifteenDaysOrLess: false,
				ExpiresInThirtyDaysOrLess:  false,
				ExpiresInSixtyDaysOrLess:   true,
			},
		},
		{
			description: "expires in 29 days",
			timeNow:     testTimeNow,
			cert: x509.Certificate{
				NotAfter: testTimeNow.AddDate(0, 0, 29),
				Subject: pkix.Name{
					CommonName: "ExpireIn29DaysCommonName",
				},
			},
			isCertValid: false,
			expetedCertErr: certError{
				CommonName:                 "ExpireIn29DaysCommonName",
				ExpirationDate:             testTimeNow.AddDate(0, 0, 29).String(),
				ExpiresInFiveDaysOrLess:    false,
				ExpiresInFifteenDaysOrLess: false,
				ExpiresInThirtyDaysOrLess:  true,
				ExpiresInSixtyDaysOrLess:   true,
			},
		},
		{
			description: "expires in 14 days",
			timeNow:     testTimeNow,
			cert: x509.Certificate{
				NotAfter: testTimeNow.AddDate(0, 0, 14),
				Subject: pkix.Name{
					CommonName: "ExpireIn14DaysCommonName",
				},
			},
			isCertValid: false,
			expetedCertErr: certError{
				CommonName:                 "ExpireIn14DaysCommonName",
				ExpirationDate:             testTimeNow.AddDate(0, 0, 14).String(),
				ExpiresInFiveDaysOrLess:    false,
				ExpiresInFifteenDaysOrLess: true,
				ExpiresInThirtyDaysOrLess:  true,
				ExpiresInSixtyDaysOrLess:   true,
			},
		},
		{
			description: "expires in 4 days",
			timeNow:     testTimeNow,
			cert: x509.Certificate{
				NotAfter: testTimeNow.AddDate(0, 0, 4),
				Subject: pkix.Name{
					CommonName: "ExpireIn4DaysCommonName",
				},
			},
			isCertValid: false,
			expetedCertErr: certError{
				CommonName:                 "ExpireIn4DaysCommonName",
				ExpirationDate:             testTimeNow.AddDate(0, 0, 4).String(),
				ExpiresInFiveDaysOrLess:    true,
				ExpiresInFifteenDaysOrLess: true,
				ExpiresInThirtyDaysOrLess:  true,
				ExpiresInSixtyDaysOrLess:   true,
			},
		},
	}

	for _, test := range tests {
		g.Describe("validateCertificate()", func() {
			g.It(fmt.Sprintf("When the cert %v, then thre will be errors is : %v", test.description, !test.isCertValid), func() {
				isCertValid, certError := validateCertificate(test.timeNow, test.cert)
				g.Assert(isCertValid).Equal(test.isCertValid)
				if !test.isCertValid {
					g.Assert(certError.CommonName).Equal(test.expetedCertErr.CommonName)
					g.Assert(certError.ExpirationDate).Equal(test.expetedCertErr.ExpirationDate)
					g.Assert(certError.ExpiresInFiveDaysOrLess).Equal(test.expetedCertErr.ExpiresInFiveDaysOrLess)
					g.Assert(certError.ExpiresInFifteenDaysOrLess).Equal(test.expetedCertErr.ExpiresInFifteenDaysOrLess)
					g.Assert(certError.ExpiresInThirtyDaysOrLess).Equal(test.expetedCertErr.ExpiresInThirtyDaysOrLess)
					g.Assert(certError.ExpiresInSixtyDaysOrLess).Equal(test.expetedCertErr.ExpiresInSixtyDaysOrLess)
				}
			})
		})
	}
}
