package sslCheck

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/franela/goblin"
)

func TestRun(t *testing.T) {
	g := goblin.Goblin(t)
	var fakeLog = logrus.New()
	var tests = []struct {
		config Config
	}{
		{
			config: Config{
				Hosts: []string{"www.usatoday.com:443"},
			},
		},
	}

	for _, test := range tests {
		g.Describe("Run()", func() {
			g.It("Run Executes without error", func() {
				Run(fakeLog, test.config, []byte{}, false, "version")
			})
		})
	}
}

func TestCheckHost(t *testing.T) {
	g := goblin.Goblin(t)

	var tests = []struct {
		description         string
		validForDays        int
		listenPort          string
		host                string
		hostHasError        bool
		shouldReturnCertErr bool
		expectedCertErr     certError
	}{
		{
			description:         "When cert is valid for 5",
			validForDays:        5,
			listenPort:          ":9443",
			host:                "127.0.0.1:9443",
			hostHasError:        false,
			shouldReturnCertErr: true,
			expectedCertErr: certError{
				ExpiresInFiveDaysOrLess:    true,
				ExpiresInFifteenDaysOrLess: true,
				ExpiresInThirtyDaysOrLess:  true,
				ExpiresInSixtyDaysOrLess:   true,
			},
		},
		{
			description:         "When cert is valid for 15",
			validForDays:        15,
			listenPort:          ":9444",
			host:                "127.0.0.1:9444",
			hostHasError:        false,
			shouldReturnCertErr: true,
			expectedCertErr: certError{
				ExpiresInFiveDaysOrLess:    false,
				ExpiresInFifteenDaysOrLess: true,
				ExpiresInThirtyDaysOrLess:  true,
				ExpiresInSixtyDaysOrLess:   true,
			},
		},
		{
			description:         "When cert is valid for 30",
			validForDays:        30,
			listenPort:          ":9445",
			host:                "127.0.0.1:9445",
			hostHasError:        false,
			shouldReturnCertErr: true,
			expectedCertErr: certError{
				ExpiresInFiveDaysOrLess:    false,
				ExpiresInFifteenDaysOrLess: false,
				ExpiresInThirtyDaysOrLess:  true,
				ExpiresInSixtyDaysOrLess:   true,
			},
		},
		{
			description:         "When cert is valid for 60",
			validForDays:        60,
			listenPort:          ":9446",
			host:                "127.0.0.1:9446",
			hostHasError:        false,
			shouldReturnCertErr: true,
			expectedCertErr: certError{
				ExpiresInFiveDaysOrLess:    false,
				ExpiresInFifteenDaysOrLess: false,
				ExpiresInThirtyDaysOrLess:  false,
				ExpiresInSixtyDaysOrLess:   true,
			},
		},
		{
			description:         "When cert is valid for 120",
			validForDays:        120,
			listenPort:          ":9447",
			host:                "127.0.0.1:9447",
			hostHasError:        false,
			shouldReturnCertErr: false,
		},
		{
			description:         "When cert is not valid",
			validForDays:        0,
			listenPort:          ":9448",
			host:                "127.0.0.1:9448",
			hostHasError:        true,
			shouldReturnCertErr: false,
		},
	}

	for _, test := range tests {
		g.Describe("ProcessHosts()", func() {
			serverCrt, CAPem := GenerateCertBundle(test.validForDays)
			g.Before(func() {
				go func() {
					config := &tls.Config{Certificates: []tls.Certificate{serverCrt}}
					ln, err := tls.Listen("tcp", test.listenPort, config)
					if err != nil {
						fmt.Printf("Error listening on port  %v err: %v", test.listenPort, err)
					}
					conn, err := ln.Accept()
					if err != nil {
						fmt.Printf("Error Accepting Conn %v", err)
					}
					conn.Write([]byte("test\n"))
					conn.Close()
					ln.Close()
				}()
			})
			g.It(test.description, func() {
				result := checkHost(test.host, CAPem)
				fmt.Printf("result %s", result)
				g.Assert(result.Err != nil).Equal(test.hostHasError)
				if test.shouldReturnCertErr {
					g.Assert(result.CertErrors[0].ExpiresInFifteenDaysOrLess).Equal(test.expectedCertErr.ExpiresInFifteenDaysOrLess)
					g.Assert(result.CertErrors[0].ExpiresInFiveDaysOrLess).Equal(test.expectedCertErr.ExpiresInFiveDaysOrLess)
					g.Assert(result.CertErrors[0].ExpiresInSixtyDaysOrLess).Equal(test.expectedCertErr.ExpiresInSixtyDaysOrLess)
					g.Assert(result.CertErrors[0].ExpiresInThirtyDaysOrLess).Equal(test.expectedCertErr.ExpiresInThirtyDaysOrLess)
				}
			})
		})
	}
}

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

// helper function to create a cert template with a serial number and other required fields
func CertTemplate(validForDays int) (*x509.Certificate, error) {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, errors.New("failed to generate serial number: " + err.Error())
	}

	tmpl := x509.Certificate{
		SerialNumber:          serialNumber,
		Subject:               pkix.Name{Organization: []string{"Gannett"}},
		SignatureAlgorithm:    x509.SHA256WithRSA,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour * 24 * time.Duration(validForDays)),
		BasicConstraintsValid: true,
	}
	return &tmpl, nil
}

func CreateCert(template, parent *x509.Certificate, pub interface{}, parentPriv interface{}) (cert *x509.Certificate, certPEM []byte, err error) {
	certDER, err := x509.CreateCertificate(rand.Reader, template, parent, pub, parentPriv)
	if err != nil {
		return
	}
	// parse the resulting certificate so we can use it again
	cert, err = x509.ParseCertificate(certDER)
	if err != nil {
		return
	}
	// PEM encode the certificate (this is a standard TLS encoding)
	b := pem.Block{Type: "CERTIFICATE", Bytes: certDER}
	certPEM = pem.EncodeToMemory(&b)
	return
}

func GenerateCertBundle(validForDays int) (tls.Certificate, []byte) {
	//Generate the Root CA PEM
	caCertTmpl, err := CertTemplate(validForDays)
	if err != nil {
		fmt.Errorf("creating cert template: %v", err)
	}
	// describe what the certificate will be used for
	caCertTmpl.IsCA = true
	caCertTmpl.KeyUsage = x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature
	caCertTmpl.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth}
	caCertTmpl.IPAddresses = []net.IP{net.ParseIP("127.0.0.1")}
	// generate a new key-pair
	caKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		fmt.Errorf("generating random key: %v", err)
	}
	CACert, CACertPEM, err := CreateCert(caCertTmpl, caCertTmpl, &caKey.PublicKey, caKey)

	//Generate the Server PEM
	if err != nil {
		fmt.Errorf("generating random key: %v", err)
	}

	// create a template for the server
	servCertTmpl, err := CertTemplate(validForDays)
	if err != nil {
		fmt.Errorf("creating cert template: %v", err)
	}
	servCertTmpl.KeyUsage = x509.KeyUsageDigitalSignature
	servCertTmpl.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}
	servCertTmpl.IPAddresses = []net.IP{net.ParseIP("127.0.0.1")}
	servKey, err := rsa.GenerateKey(rand.Reader, 2048)
	_, servCertPEM, err := CreateCert(servCertTmpl, CACert, &servKey.PublicKey, caKey)

	servKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(servKey),
	})
	serverTLSCert, err := tls.X509KeyPair(servCertPEM, servKeyPEM)
	if err != nil {
		fmt.Errorf("invalid key pair: %v", err)
	}
	return serverTLSCert, CACertPEM
}
