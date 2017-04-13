package sslCheck

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/GannettDigital/go-newrelic-plugin/helpers"
	"github.com/Sirupsen/logrus"
)

//Config is the keeper of the config
type Config struct {
	Hosts              []string
	ExpiredEventPeriod int
}

type certError struct {
	CommonName                 string
	ExpirationDate             string
	ExpiresInFiveDaysOrLess    bool
	ExpiresInFifteenDaysOrLess bool
	ExpiresInThirtyDaysOrLess  bool
	ExpiresInSixtyDaysOrLess   bool
}

type hostResult struct {
	Host       string
	Err        error
	CertErrors []certError
}

// InventoryData is the data type for inventory data produced by a plugin data
// source and emitted to the agent's inventory data store
type inventoryData map[string]interface{}

// MetricData is the data type for events produced by a plugin data source and
// emitted to the agent's metrics data store
type metricData map[string]interface{}

// EventData is the data type for single shot events
type eventData map[string]interface{}

// PluginData defines the format of the output JSON that plugins will return
type pluginData struct {
	Name            string                   `json:"name"`
	ProtocolVersion string                   `json:"protocol_version"`
	PluginVersion   string                   `json:"plugin_version"`
	Metrics         []metricData             `json:"metrics"`
	Inventory       map[string]inventoryData `json:"inventory"`
	Events          []eventData              `json:"events"`
	Status          string                   `json:"status"`
}

const NAME string = "sslCheck"
const EVENT_TYPE_VALID string = "GSSLSampleValid"
const EVENT_TYPE_INVALID string = "GSSLSampleInvalid"
const PROVIDER string = "sslChecker"
const PROTOCOL_VERSION string = "1"

const FiveDays = 5
const FifteenDays = 15
const ThirtyDays = 30
const SixtyDays = 60

func Run(log *logrus.Logger, config Config, prettyPrint bool, version string) {
	// Initialize the output structure
	var data = pluginData{
		Name:            NAME,
		ProtocolVersion: PROTOCOL_VERSION,
		PluginVersion:   version,
		Inventory:       make(map[string]inventoryData),
		Metrics:         make([]metricData, 0),
		Events:          make([]eventData, 0),
	}

	for _, host := range config.Hosts {
		result := checkHost(host)
		if result.Err != nil {
			data.Metrics = append(data.Metrics, map[string]interface{}{
				"event_type": EVENT_TYPE_INVALID,
				"provider":   PROVIDER,
				"host":       result.Host,
				"reason":     result.Err.Error(),
			})
		}
		if len(result.CertErrors) > 0 {
			for _, certError := range result.CertErrors {
				data.Metrics = append(data.Metrics, map[string]interface{}{
					"event_type":      EVENT_TYPE_VALID,
					"provider":        PROVIDER,
					"host":            certError.CommonName,
					"expirationDate":  certError.ExpirationDate,
					"expiresIn60Days": certError.ExpiresInSixtyDaysOrLess,
					"expiresIn30Days": certError.ExpiresInThirtyDaysOrLess,
					"expiresIn15Days": certError.ExpiresInFifteenDaysOrLess,
					"expiresIn5Days":  certError.ExpiresInFiveDaysOrLess,
				})
			}
		}
	}

	fatalIfErr(log, helpers.OutputJSON(data, prettyPrint))
}

func ValidateConfig(config Config) error {
	if len(config.Hosts) == 0 {
		return errors.New("You must provide hosts to check")
	}
	return nil
}

func fatalIfErr(log *logrus.Logger, err error) {
	if err != nil {
		log.WithError(err).Fatal("can't continue")
	}
}

// ProcessHosts processes a string of hosts seperated by comma.
// for example: "example.com:443,google.com:443,gcipaas.com:8443"
// Ensure the listening port is included.
// Host is invalid if the port is not included and will not be included in the list of hosts.
func ProcessHosts(hosts string) ([]string, error) {
	var validHosts []string
	if hosts == "" {
		return validHosts, errors.New("You must provide hosts to check")
	}
	for _, currentHost := range strings.Split(hosts, ",") {
		if validHost(strings.Trim(currentHost, " ")) {
			validHosts = append(validHosts, currentHost)
		}
	}
	return validHosts, nil
}

func validHost(host string) bool {
	return regexp.MustCompile(`[\w\.]+:\d{1,5}`).Match([]byte(host))
}

func checkHost(host string) (result hostResult) {
	result = hostResult{
		Host:       host,
		CertErrors: []certError{},
	}
	conn, err := tls.Dial("tcp", host, nil)
	if err != nil {
		result.Err = err
		return
	}
	defer conn.Close()
	timeNow := time.Now()
	checkedCerts := make(map[string]bool)
	for _, chain := range conn.ConnectionState().VerifiedChains {
		for _, cert := range chain {
			if _, checked := checkedCerts[string(cert.Signature)]; checked {
				continue
			}
			checkedCerts[string(cert.Signature)] = true
			isCertValid, certError := validateCertificate(timeNow, *cert)
			if isCertValid {
				continue
			} else {
				result.CertErrors = append(result.CertErrors, certError)
			}
		}
	}
	return result
}

func validateCertificate(timeNow time.Time, cert x509.Certificate) (bool, certError) {
	var isCertValid bool = true
	var certError certError
	if timeNow.AddDate(0, 0, FiveDays).After(cert.NotAfter) {
		isCertValid = false
		certError.ExpiresInFiveDaysOrLess = true
	}
	if timeNow.AddDate(0, 0, FifteenDays).After(cert.NotAfter) {
		isCertValid = false
		certError.ExpiresInFifteenDaysOrLess = true
	}
	if timeNow.AddDate(0, 0, ThirtyDays).After(cert.NotAfter) {
		isCertValid = false
		certError.ExpiresInThirtyDaysOrLess = true
	}
	if timeNow.AddDate(0, 0, SixtyDays).After(cert.NotAfter) {
		isCertValid = false
		certError.ExpiresInSixtyDaysOrLess = true
	}
	if !isCertValid {
		certError.CommonName = cert.Subject.CommonName
		certError.ExpirationDate = cert.NotAfter.String()
	}
	return isCertValid, certError
}
