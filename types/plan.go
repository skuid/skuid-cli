package types

type Plan struct {
	Host     string   `json:"host"`
	Port     string   `json:"port"`
	URL      string   `json:"url"`
	Type     string   `json:"type"`
	Metadata Metadata `json:"metadata"`
}

type Metadata struct {
	Pages        []string `json:"pages"`
	Apps         []string `json:"apps"`
	DataServices []string `json:"dataservices"`
	DataSources  []string `json:"datasources"`
	Profiles     []string `json:"profiles"`
}
