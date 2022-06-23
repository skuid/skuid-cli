package pkg

import (
	"fmt"

	"github.com/valyala/fasthttp"
)

const (
	ZIP_CONTENT_TYPE  = "application/zip"
	JSON_CONTENT_TYPE = "application/json"

	DEFAULT_CONTENT_TYPE = ZIP_CONTENT_TYPE

	HEADER_SKUID_PUBLIC_KEY_ENDPOINT = "x-skuid-public-key-endpoint"
)

type RequestHeaders map[string]string

// easy macro for the authorization headers we want
func GenerateHeaders(host, token string) RequestHeaders {
	return RequestHeaders{
		fasthttp.HeaderAuthorization:     fmt.Sprintf("Bearer %v", token),
		HEADER_SKUID_PUBLIC_KEY_ENDPOINT: fmt.Sprintf("%v/api/v1/site/verificationkey", host),
	}
}

// GeneratePlanHeaders is a lot like GenerateRoute. We check whether it's a warden
// or a pliny request, then change the parameters depending on that.
func GeneratePlanHeaders(info *Authorization, plan NlxPlan) (headers RequestHeaders) {
	// warden requests all have a different host than the one we originall authenticated
	// with.
	wardenRequest := plan.Host != ""

	// when given a warden request we need to provide the authorization / jwt token
	if wardenRequest {
		headers = GenerateHeaders(plan.Host, info.AuthorizationToken)
	} else {
		headers = GenerateHeaders(plan.Host, info.AccessToken)
	}

	return
}
