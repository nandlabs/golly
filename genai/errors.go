package genai

import "oss.nandlabs.io/golly/errutils"

var GetUnsupportedModelErr = errutils.NewCustomError("unsupported model %s")
var GetUnsupportedMimeErr = errutils.NewCustomError("unsupported mime type %s for model %s")
var GetUnsupportedConsumerErr = errutils.NewCustomError("unsupported consumer for model %s")
var GetUnsupportedProviderErr = errutils.NewCustomError("unsupported provider for model %s")
var GetInvalidOptionErr = errutils.NewCustomError("invalid option %s for model %s")
