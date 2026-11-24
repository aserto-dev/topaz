package errs

type BuiltinError string

func (e BuiltinError) Error() string {
	return string(e)
}

const (
	ErrAuthZENPluginDisabled = BuiltinError("authzen plugin is disabled")
	ErrTopazPluginDisabled   = BuiltinError("topaz plugin is disabled")
	ErrInvalidDuration       = BuiltinError("invalid duration format")
)
