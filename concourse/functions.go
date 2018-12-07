package concourse

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/mitchellh/colorstring"
)

func Check(step string, err error) {
	//TODO: refactor this function to return something instead of exiting
	if err != nil {
		Fatal(step+" step failed", err)
	}
}

func Fatal(doing string, err error) {
	Sayf(colorstring.Color("[red]error %s: %s\n"), doing, err)
	//TODO: don't exit here, let the caller decide.
	os.Exit(1)
}

func Sayf(message string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, message, args...)
}

//TODO refactor this and don't exit in this function, instead return control to caller with err msg
func ReadRequest(request interface{}) {
	if err := json.NewDecoder(os.Stdin).Decode(request); err != nil {
		Fatal("Error reading request: %v\n", err)
	}
}

//TODO refactor this and don't exit in this function, instead return control to caller with err msg
func WriteResponse(response interface{}) {
	if err := json.NewEncoder(os.Stdout).Encode(response); err != nil {
		Fatal("Error writing response: %v\n", err)
	}
	os.Exit(0)
}
