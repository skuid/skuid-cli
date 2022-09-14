package lit

import (
	"fmt"

	"github.com/skuid/domain/constants"

	"github.com/skuid/tides/ui/style"
)

var (
	MainHeader = fmt.Sprintf(`Welcome to Skuid's Command Line Interface (CLI): %v %v`,
		style.StyleTides.Render("Tides"),
		style.StyleSubtle.Render(fmt.Sprintf("(version %v)", constants.VERSION_NAME)))
)
