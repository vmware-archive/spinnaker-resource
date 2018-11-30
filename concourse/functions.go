package concourse

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/mitchellh/colorstring"
)

func Check(step string, err error) {
	if err != nil {
		Fatal(step+" step failed", err)
	}
}

func Fatal(doing string, err error) {
	Sayf(colorstring.Color("[red]error %s: %s\n"), doing, err)
	os.Exit(1)
}

func Sayf(message string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, message, args...)
}

func ReadRequest(request interface{}) {
	if err := json.NewDecoder(os.Stdin).Decode(request); err != nil {
		Fatal("Error reading request: %v\n", err)
	}
}

func WriteResponse(response interface{}) {
	if err := json.NewEncoder(os.Stdout).Encode(response); err != nil {
		Fatal("Error writing response: %v\n", err)
	}
	os.Exit(0)
}
