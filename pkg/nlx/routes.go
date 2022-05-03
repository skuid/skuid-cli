package nlx

import "fmt"

// GenerateRoute is similar to GenerateHeaders. We basically just check
// whether or not something is a pliny or a warden request
func GenerateRoute(info *Authorization, plan NlxPlan) (url string) {
	// warden requests all have a different host than the one we originall authenticated
	// with.
	wardenRequest := plan.Host != ""

	// for legibility
	plinyRequest := !wardenRequest

	// when given a warden request we have to use the plan information
	// for the url
	if wardenRequest {
		if plan.Port != "" {
			url = fmt.Sprintf("%s:%s/api/%v%s", plan.Host, plan.Port, DEFAULT_API_VERSION, plan.Endpoint)
		} else {
			url = fmt.Sprintf("%s/api/%v%s", plan.Host, DEFAULT_API_VERSION, plan.Endpoint)
		}
	}

	// with a pliny request we just attach the endpoint to the same
	// host we've been using since it's a relative route
	if plinyRequest {
		url = fmt.Sprintf("%s/api/%v%s", info.Host, DEFAULT_API_VERSION, plan.Endpoint)
	}

	return
}
