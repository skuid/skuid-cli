package flags

var (
	// PlatformLoginFlags adds the required necessary flags to a command
	// for the function PlatformLogin
	PlatformLoginFlags = []*Flag[string]{
		Host, Username, Password,
	}
)
