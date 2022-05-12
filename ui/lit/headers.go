package lit

import (
	"fmt"

	"github.com/skuid/tides/pkg/constants"
	"github.com/skuid/tides/ui/style"
)

var (
	MainHeader = fmt.Sprintf(`Welcome to Skuid's Command Line Interface (CLI): %v %v`,
		style.Tides("Tides"),
		style.Subtle(fmt.Sprintf("(version %v)", constants.VERSION_NAME)))
)
