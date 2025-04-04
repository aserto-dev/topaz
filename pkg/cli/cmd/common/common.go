package common

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
)

type SaveContext bool

const (
	CLIConfigurationFile = "topaz.json"
)

var (
	Save                  SaveContext
	RestrictedNamePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_.-]*$`)
)

func PromptYesNo(label string, def bool) bool {
	choices := "Y/n"
	if !def {
		choices = "y/N"
	}

	sigChan := make(chan os.Signal, 1)
	defer close(sigChan)

	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	inputChan := make(chan string)
	defer close(inputChan)

	fmt.Fprintf(os.Stderr, "%s (%s) ", label, choices)

	go func() {
		reader := bufio.NewReader(os.Stdin)

		for {
			text, _ := reader.ReadString('\n')
			inputChan <- strings.TrimSpace(text)
		}
	}()

	for {
		select {
		case input := <-inputChan:
			switch {
			case input == "Y" || input == "y":
				return true
			case input == "N" || input == "n":
				return false
			case input == "":
				return def
			}

		case <-sigChan:
			return false
		}

		fmt.Fprintf(os.Stderr, "%s (%s) ", label, choices)
	}
}
