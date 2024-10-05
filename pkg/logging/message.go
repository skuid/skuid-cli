package logging

import (
	"fmt"
	"iter"
	"slices"
	"strconv"
	"strings"

	"github.com/bobg/seqs"
	"github.com/gookit/color"
)

type LogColor struct {
	color.Color
}

func (c *LogColor) QuoteText(message any) string {
	return c.Text(QuoteText(message))
}

var (
	ColorEnvVar   = LogColor{color.Blue}    // Environment variables
	ColorResource = LogColor{color.Cyan}    // Client or server site resources (e.g., site host, directory/file names, entity names, etc.)
	ColorFilter   = LogColor{color.Magenta} // Resources that are filtered in or out (e.g., entity names, app name, since value, metadata type names, etc.)
	ColorSuccess  = LogColor{color.Green}   // Successful tracked activity
	ColorFailure  = LogColor{color.Red}     // Failed tracked activity
	ColorWarning  = LogColor{color.Yellow}  // Call attention to user of a significant event and/or condition but processing is continuing
	ColorStart    = LogColor{color.Gray}    // Starting tracked activity
)

const (
	SuccessText     = "SUCCESS"
	FailureText     = "FAILURE"
	StartText       = "START"
	SuccessIcon     = "✓"
	FailureIcon     = "X"
	FileAnIssueText = "please try again and if the problem persists create an issue at https://github.com/skuid/skuid-cli/issues"
)

func ColorSuccessIcon() string {
	return ColorSuccess.Text(SuccessIcon)
}

func ColorFailureIcon() string {
	return ColorFailure.Text(FailureIcon)
}

func ColorStartText() string {
	return ColorStart.Text(StartText)
}

func ColorResult(err error) LogColor {
	return ColorResultCondition(err == nil)
}

func ColorResultCondition(condition bool) LogColor {
	if condition {
		return ColorSuccess
	} else {
		return ColorFailure
	}
}

func ColorResultText(err error) string {
	return ColorResultConditionText(err == nil)
}

func ColorResultConditionText(condition bool) string {
	c := ColorResultCondition(condition)
	if condition {
		return c.Text(SuccessText)
	} else {
		return c.Text(FailureText)
	}
}

// Returns the value wrapped with back-ticks
// Equivalent to calling fmt.Sprintf("%#q", val)
// Use back-tick instead of quote because log files quote all fields in the file
// which makes the logs incredibly difficult to read/parse because of all the escape
// characters and in certain situations depending on the characters contained in the
// value itself, wrecks havoc on escaping escape characters that may be present in
// that value.
//
// Should be used whenever a logged value requires quotes around it instead of applying
// the quotes within the formatted message for a couple of reasons:
//  1. Ensures that all values in the logs are quoted the same way
//  2. Allows for easily changing the quote character if needed
func QuoteText(val any) string {
	// sprintf is expensive so short-circuit if possible
	// as currently written, nothing should have backticks unless we put them
	// there and the only thing that will likely run through this code is Path
	// names (directories and files) so theoertically could just add the backtick
	// and be done with it but just in case, we check to make sure we're able
	// before passing through the Sprintf to fully resolve if not
	s, ok := val.(string)
	if ok && strconv.CanBackquote(s) {
		return "`" + s + "`"
	} else {
		return fmt.Sprintf("%#q", val)
	}
}

func QuoteItemText[T any](inp iter.Seq[T]) iter.Seq[string] {
	return seqs.Map(inp, func(val T) string {
		return QuoteText(val)
	})
}

// Returns a sorted list in the form "[item1, item2]" with the list optionally wrapped in a
// color and each item is separated by delimiter and optionally wrapped with a back-tick
func SortedDelimitedList[T any](inp iter.Seq[T], quoted bool, c *LogColor, delimiter string) string {
	s := "[" + strings.Join(slices.Sorted(QuoteItemText(inp)), delimiter) + "]"
	if c != nil {
		return c.Text(s)
	} else {
		return s
	}
}

// Returns a sorted list in CSV format in the form "[`item`, `item2`]". Equivalent to calling
// fmt.Sprintf("%#q", slice) but with comma delimited items.
func CSV[T any](inp iter.Seq[T]) string {
	return SortedDelimitedList(inp, true, nil, ", ")
}

// Returns a sorted list in CSV format in the form "[`item`, `item2`]" with the list optionally
// wrapped in a color. Equivalent to calling fmt.Sprintf("%#q", slice) but with comma delimited items.
func CSVColor[T any](inp iter.Seq[T], c *LogColor) string {
	return SortedDelimitedList(inp, true, c, ", ")
}
