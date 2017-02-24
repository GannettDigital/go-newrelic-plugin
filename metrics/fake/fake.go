package fake

import (
	"net/http"

	"github.com/Sirupsen/logrus"
	newrelic "github.com/newrelic/go-agent"
)

// HTTPResult - implementes APIRunner
type HTTPResult struct {
	Code int
	Data []byte
	Err  error
}

// CallAPI -
func (fetcher HTTPResult) CallAPI(*logrus.Logger, newrelic.Transaction, *http.Request, *http.Client) (int, []byte, error) {
	return fetcher.Code, fetcher.Data, fetcher.Err
}
