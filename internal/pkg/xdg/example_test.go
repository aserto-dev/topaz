package xdg_test

import (
	"fmt"

	"github.com/aserto-dev/topaz/internal/pkg/xdg"
)

func ExampleDataFile() {
	dataFilePath, err := xdg.DataFile("appname/app.data")
	if err != nil {
		// Treat error.
	}

	fmt.Println("Save data file at:", dataFilePath)
}

func ExampleConfigFile() {
	configFilePath, err := xdg.ConfigFile("appname/app.yaml")
	if err != nil {
		// Treat error.
	}

	fmt.Println("Save config file at:", configFilePath)
}

func ExampleStateFile() {
	stateFilePath, err := xdg.DataFile("appname/app.state")
	if err != nil {
		// Treat error.
	}

	fmt.Println("Save state file at:", stateFilePath)
}

func ExampleCacheFile() {
	cacheFilePath, err := xdg.CacheFile("appname/app.cache")
	if err != nil {
		// Treat error.
	}

	fmt.Println("Save cache file at:", cacheFilePath)
}

func ExampleRuntimeFile() {
	runtimeFilePath, err := xdg.RuntimeFile("appname/app.pid")
	if err != nil {
		// Treat error.
	}

	fmt.Println("Save runtime file at:", runtimeFilePath)
}

func ExampleSearchDataFile() {
	dataFilePath, err := xdg.SearchDataFile("appname/app.data")
	if err != nil {
		// The data file could not be found.
	}

	fmt.Println("The data file was found at:", dataFilePath)
}

func ExampleSearchConfigFile() {
	configFilePath, err := xdg.SearchConfigFile("appname/app.yaml")
	if err != nil {
		// The config file could not be found.
	}

	fmt.Println("The config file was found at:", configFilePath)
}

func ExampleSearchStateFile() {
	stateFilePath, err := xdg.SearchStateFile("appname/app.state")
	if err != nil {
		// The state file could not be found.
	}

	fmt.Println("The state file was found at:", stateFilePath)
}

func ExampleSearchCacheFile() {
	cacheFilePath, err := xdg.SearchCacheFile("appname/app.cache")
	if err != nil {
		// The cache file could not be found.
	}

	fmt.Println("The cache file was found at:", cacheFilePath)
}

func ExampleSearchRuntimeFile() {
	runtimeFilePath, err := xdg.SearchRuntimeFile("appname/app.pid")
	if err != nil {
		// The runtime file could not be found.
	}

	fmt.Println("The runtime file was found at:", runtimeFilePath)
}
