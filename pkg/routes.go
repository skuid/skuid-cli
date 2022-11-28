package pkg

import "fmt"

// GenerateRoute is similar to GenerateHeaders. We basically just check
// whether or not something is a pliny or a warden request
func GenerateRoute(info *Authorization, plan NlxPlan) (url string) {
	// pliny requests will be to the same host that info authorization came from,
	// so plan.Host will be empty string
	wardenRequest := plan.Host != ""

	// when given a warden request we have to use the plan information
	// for the url
	if wardenRequest {
		if plan.Port != "" {
			url = fmt.Sprintf("%s:%s/api/%v%s", plan.Host, plan.Port, DEFAULT_API_VERSION, plan.Endpoint)
		} else {
			url = fmt.Sprintf("%s/api/%v%s", plan.Host, DEFAULT_API_VERSION, plan.Endpoint)
		}
	} else /* pliny request */ {
		url = fmt.Sprintf("%s/api/%v%s", info.Host, DEFAULT_API_VERSION, plan.Endpoint)
	}

	return
}
