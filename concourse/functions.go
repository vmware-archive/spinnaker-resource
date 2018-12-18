/*
Copyright (C) 2018-Present Pivotal Software, Inc. All rights reserved.

This program and the accompanying materials are made available under the terms of the under the Apache License, Version 2.0 (the "License”); you may not use this file except in compliance with the License. You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.
*/
package concourse

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/mitchellh/colorstring"
)

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
