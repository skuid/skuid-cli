package types

type RetrieveRequest struct {
	Metadata map[string]map[string]string `json:metadata`
}

var MetadataTypes map[string]string = map[string]string{
	"apps":          "apps",
	"authproviders": "authProviders",
	"datasources":   "dataSources",
	"pages":         "pages",
	"profiles":      "profiles",
	"themes":        "themes",
}
