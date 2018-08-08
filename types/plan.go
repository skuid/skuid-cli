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

func (m Metadata) GetNamesForType(metadataType string) []string {
	switch metadataType {
	case "pages":
		return m.Pages
	case "apps":
		return m.Apps
	case "dataservices":
		return m.DataServices
	case "datasources":
		return m.DataSources
	case "profiles":
		return m.Profiles
	default:
		return nil
	}
}
