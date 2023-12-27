package spanlint

// Config is a configuration for the spanlint analyzer.
type Config struct {
	// EnableEndCheck enables the check for calling span.End().
	DisableEndCheck bool

	// EnableSetStatusCheck enables the check for calling span.SetStatus.
	EnableSetStatusCheck bool

	// EnableRecordErrorCheck enables the check for calling span.RecordError.
	// By default, this check is disabled.
	EnableRecordErrorCheck bool
}

// DefaultConfig is the default configuration for the spanlint analyzer.
var DefaultConfig = &Config{
	DisableEndCheck:        false,
	EnableSetStatusCheck:   false,
	EnableRecordErrorCheck: false,
}
