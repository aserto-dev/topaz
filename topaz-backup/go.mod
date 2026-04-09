module github.com/aserto-dev/topaz/topaz-backup

go 1.26.1

replace github.com/aserto-dev/topaz/internal => ../internal

replace github.com/aserto-dev/topaz/topaz => ../topaz

require (
	github.com/alecthomas/kong v1.15.0
	github.com/aserto-dev/topaz/internal v0.0.0-00010101000000-000000000000
	github.com/pkg/errors v0.9.1
	go.etcd.io/bbolt v1.4.3
)

require golang.org/x/sys v0.43.0 // indirect
