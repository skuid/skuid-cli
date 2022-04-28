package flags

var (
	Modules = &Flag[[]string]{
		argument:  &argModules,
		Name:      "modules",
		Shorthand: "m",
		Usage:     "Module name(s), separated by a comma.",
	}

	Pages = &Flag[[]string]{
		argument:  &argPages,
		Name:      "pages",
		Shorthand: "n",
		Usage:     "Page name(s), separated by a comma.",
	}
)
