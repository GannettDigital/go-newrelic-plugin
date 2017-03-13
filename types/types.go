package types

import (
	"github.com/Sirupsen/logrus"
)

type Opts struct {
	Type        string `long:"type" description:"The type of collector to run. Needs to match as defined in CollectorArray"`
	Verbose     bool   `long:"verbose" description:"Print more information to logs"`
	PrettyPrint bool   `long:"pretty-print" description:"Print pretty formatted JSON"`
	Version     bool   `long:"version" description:"Print version information and exit"`
}

type Collector func(log *logrus.Logger, opts Opts, version string)