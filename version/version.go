package version

import (
	"fmt"
	"io/ioutil"
)

// Name of this version of Skuid CLI
var Name = "unknown"

func init() {
	version, err := ioutil.ReadFile(".version")
	if err != nil {
		fmt.Printf("Could not read CLI version from .version: %s\n", err.Error())
		return
	}
	Name = string(version)
}
