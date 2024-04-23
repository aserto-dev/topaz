package resolvers

type Resolvers struct {
	runtimeResolver   RuntimeResolver
	directoryResolver DirectoryResolver
	edgeSyncResolver  EdgeSyncResolver
}

func New() *Resolvers {
	return &Resolvers{}
}

func (s *Resolvers) SetRuntimeResolver(resolver RuntimeResolver) {
	s.runtimeResolver = resolver
}

func (s *Resolvers) GetRuntimeResolver() RuntimeResolver {
	return s.runtimeResolver
}

func (s *Resolvers) SetDirectoryResolver(resolver DirectoryResolver) {
	s.directoryResolver = resolver
}

func (s *Resolvers) GetDirectoryResolver() DirectoryResolver {
	return s.directoryResolver
}

func (s *Resolvers) SetEdgeSyncResolver(resolver EdgeSyncResolver) {
	s.edgeSyncResolver = resolver
}

func (s *Resolvers) GetEdgeSyncResolver() EdgeSyncResolver {
	return s.edgeSyncResolver
}
