package cmdutil

import (
	"github.com/skuid/skuid-cli/pkg/flags"
	"github.com/spf13/pflag"
)

// environment variables are flag specific so we track each flag created
// and then lookup its corresponding environment variable.  For example, if a new command
// was added (e.g., dosomething) which had a flag of --dir, it should not map to the legacy
// SKUID_DEFAULT_FOLDER environment variable since it never supported it.
type flagMap map[*pflag.Flag]flags.FlagInfo

type FlagTracker interface {
	Track(f *pflag.Flag, r flags.FlagInfo)
}

type FlagFinder interface {
	Find(pf *pflag.Flag) (flags.FlagInfo, bool)
}

type FlagTrackerFinder interface {
	FlagTracker
	FlagFinder
}

type CommandFlags struct {
	Int            []*flags.Flag[int]
	String         []*flags.Flag[string]
	RedactedString []*flags.Flag[flags.RedactedString]
	Bool           []*flags.Flag[bool]
	StringSlice    []*flags.Flag[flags.StringSlice]
	CustomString   []*flags.Flag[flags.CustomString]
}

type FlagManager struct {
	flags flagMap
}

func (fm *FlagManager) Track(f *pflag.Flag, r flags.FlagInfo) {
	if fm.flags == nil {
		fm.flags = make(flagMap)
	}
	fm.flags[f] = r
}

func (fm *FlagManager) Find(pf *pflag.Flag) (flags.FlagInfo, bool) {
	f, ok := fm.flags[pf]
	return f, ok
}

func NewFlagManager() *FlagManager {
	return &FlagManager{}
}
