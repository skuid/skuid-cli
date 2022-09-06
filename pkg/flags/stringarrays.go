package flags

var (
	Modules = &Flag[[]string]{
		Name:      "modules",
		Shorthand: "m",
		Usage:     "Module name(s), separated by a comma",
	}

	Pages = &Flag[[]string]{
		Name:      "pages",
		Shorthand: "n",
		Usage:     "Page name(s), separated by a comma",
	}
)
