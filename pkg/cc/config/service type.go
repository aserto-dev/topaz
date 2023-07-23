package config

type ServiceType int

// Known services that can be configured.
const (
	unknown ServiceType = iota
	authorizer
	reader
	writer
	importer
	exporter
)

func (s ServiceType) String() string {
	switch s {
	case authorizer:
		return "authorizer"
	case reader:
		return "reader"
	case writer:
		return "writer"
	case importer:
		return "importer"
	case exporter:
		return "exporter"
	case unknown:
		return "unknown"
	}

	return ""
}

var ServiceTypeMap = func() map[string]ServiceType {
	serviceMap := make(map[string]ServiceType)
	for i := authorizer; i <= exporter; i++ {
		serviceMap[i.String()] = i
	}
	return serviceMap
}
