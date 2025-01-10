package messaging

import (
	"oss.nandlabs.io/golly/l3"
)

// Logger for this package
var logger = l3.Get()

func init() {
	GetManager()
}
