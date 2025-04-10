package builder

type ServiceTypes interface {
	AvailableServices() []string
	GetGRPCRegistrations(services ...string) GRPCRegistrations
	GetGatewayRegistration(port string, services ...string) HandlerRegistrations
	Cleanups() []func()
}
