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

func ReadCheckRequest(request *CheckRequest) {
	if err := json.NewDecoder(os.Stdin).Decode(request); err != nil {
		Fatal("Error reading request: %v\n", err)
	}
}

func ReadInRequest(request *InRequest) {
	if err := json.NewDecoder(os.Stdin).Decode(request); err != nil {
		Fatal("Error reading request: %v\n", err)
	}
}

//TODO remove from commons, rename
func ReadRequest(request *OutRequest) {
	if err := json.NewDecoder(os.Stdin).Decode(request); err != nil {
		Fatal("Error reading request: %v\n", err)
	}
}

func WriteCheckResponse(response CheckResponse) {
	if err := json.NewEncoder(os.Stdout).Encode(response); err != nil {
		Fatal("Error writing response: %v\n", err)
	}
}

func WriteResponse(response OutResponse) {
	if err := json.NewEncoder(os.Stdout).Encode(response); err != nil {
		Fatal("Error writing response: %v\n", err)
	}
}

func WriteInResponse(response InResponse) {
	if err := json.NewEncoder(os.Stdout).Encode(response); err != nil {
		Fatal("Error writing response: %v\n", err)
	}
}
