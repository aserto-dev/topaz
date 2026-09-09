package errs

type Error string

func (e Error) Error() string {
	return string(e)
}

const (
	ErrAuthZENPluginDisabled = Error("authzen plugin is disabled")
	ErrTopazPluginDisabled   = Error("topaz plugin is disabled")
	ErrInvalidDuration       = Error("invalid duration format")
)
