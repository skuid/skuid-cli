package flags

var (
	// NLXLoginFlags adds the required necessary flags to a command
	// for the function NLXLogin
	NLXLoginFlags = []*Flag[string]{
		PlinyHost, Username, Password,
	}
)
