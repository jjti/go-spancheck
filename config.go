package spanlint

import "regexp"

// Config is a configuration for the spanlint analyzer.
type Config struct {
	// EnableEndCheck enables the check for calling span.End().
	DisableEndCheck bool

	// EnableSetStatusCheck enables the check for calling span.SetStatus.
	EnableSetStatusCheck bool

	// IgnoreSetStatusCheckSignatures is a regex that, if matched, disables the
	// SetStatus check for a particular error.
	IgnoreSetStatusCheckSignatures *regexp.Regexp

	// EnableRecordErrorCheck enables the check for calling span.RecordError.
	// By default, this check is disabled.
	EnableRecordErrorCheck bool

	// IgnoreRecordErrorCheckSignatures is a regex that, if matched, disables the
	// RecordError check for a particular error.
	IgnoreRecordErrorCheckSignatures *regexp.Regexp
}

// NewDefaultConfig returns a new Config with default values.
func NewDefaultConfig() *Config {
	return &Config{
		DisableEndCheck:                  false,
		EnableSetStatusCheck:             false,
		IgnoreSetStatusCheckSignatures:   nil,
		EnableRecordErrorCheck:           false,
		IgnoreRecordErrorCheckSignatures: nil,
	}
}
