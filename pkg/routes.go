package pkg

import (
	"fmt"

	"github.com/skuid/skuid-cli/pkg/logging"
)

// GenerateRoute is similar to GenerateHeaders. We basically just check
// whether or not something is a pliny or a warden request
func GeneratePlanRoute(info *Authorization, planName PlanName, host string, port string, endpoint string) string {
	// pliny requests will be to the same host that info authorization came from,
	// so plan.Host will be empty string. When given a warden request we have to
	// use the plan informationfor the url
	if planName == PlanNameWarden {
		if port != "" {
			port = ":" + port
		}
	} else if planName == PlanNamePliny {
		host = info.Host
		// Skuid Review Required - The v0.6.7 code would ignore any port on the Plan that was passed in when generating the URL
		// Assuming it is expected that Pliny would never have a port but it seems odd that if the plan had a port it would be
		// ignored rather than factored in.  Is there a specific reason, other than "there shouldn't be one so let's just ignore
		// it if there is" that it ignored the port or should the port be applied if in fact the Plan returned from the server
		// contains a port in the response?  For example, if the server side ever changes and ports start being used, the CLI
		// would need to be updated - seems more appropriate to just use the value that the server gave us since it "should"
		// be returning proper values for host/port/etc.
		//
		// TODO: Adjust port based on answer to above
		port = ""
	} else {
		// should not happen in production
		panic(fmt.Errorf("unexpected plan name %v", logging.QuoteText(planName)))
	}
	return fmt.Sprintf("%s%s/api/%v%s", host, port, DefaultApiVersion, endpoint)
}
