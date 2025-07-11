package x

import "os"

const (
	FileMode0600 os.FileMode = 0o600 // drw-------
	FileMode0700 os.FileMode = 0o700 // drwx------
	FileMode0755 os.FileMode = 0o755 // drwxr-xr-x
)
