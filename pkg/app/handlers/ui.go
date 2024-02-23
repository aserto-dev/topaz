package handlers

import (
	"fmt"
	"net/http"
	"strings"
)

type fsWithDefinition struct {
	consoleFS http.FileSystem
}

func (f *fsWithDefinition) Open(name string) (http.File, error) {
	if strings.HasPrefix(name, "/ui/") {
		return f.consoleFS.Open("console/index.html")
	}

	name = strings.TrimPrefix(name, "/public")
	return f.consoleFS.Open(fmt.Sprintf("console%s", name))
}

func UIHandler(consoleFS http.FileSystem) http.Handler {
	return http.FileServer(&fsWithDefinition{consoleFS: consoleFS})
}
