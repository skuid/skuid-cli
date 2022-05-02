package nlx

import "encoding/json"

type NlxPlan struct {
	Host     string      `json:"host"`
	Port     string      `json:"port"`
	Url      string      `json:"url"`
	Type     string      `json:"type"`
	Metadata NlxMetadata `json:"metadata"`
	Warnings []string    `json:"warnings"`
}

type NlxMetadata struct {
	Apps               []string `json:"apps"`
	AuthProviders      []string `json:"authproviders"`
	ComponentPacks     []string `json:"componentpacks"`
	DataServices       []string `json:"dataservices"`
	DataSources        []string `json:"datasources"`
	DesignSystems      []string `json:"designsystems"`
	Variables          []string `json:"variables"`
	Files              []string `json:"files"`
	Pages              []string `json:"pages"`
	PermissionSets     []string `json:"permissionsets"`
	SitePermissionSets []string `json:"sitepermissionsets"`
	Site               []string `json:"site"`
	Themes             []string `json:"themes"`
}

// for backwards compatibility
func (m *NlxMetadata) UnmarshalJSON(data []byte) error {
	// unmarshal the old fields
	old := struct {
		Apps           []string `json:"apps"`
		AuthProviders  []string `json:"authproviders"`
		ComponentPacks []string `json:"componentpacks"`
		DataServices   []string `json:"dataservices"`
		DataSources    []string `json:"datasources"`
		DesignSystems  []string `json:"designsystems"`
		Variables      []string `json:"variables"`
		Files          []string `json:"files"`
		Pages          []string `json:"pages"`
		PermissionSets []string `json:"permissionsets"`
		Profiles       []string `json:"profiles"`
		Site           []string `json:"site"`
		Themes         []string `json:"themes"`
	}{}
	err := json.Unmarshal(data, &old)
	if err != nil {
		return err
	} else {
		m.Apps = old.Apps
		m.AuthProviders = old.AuthProviders
		m.ComponentPacks = old.ComponentPacks
		m.DataServices = old.DataServices
		m.DataSources = old.DataSources
		m.DesignSystems = old.DesignSystems
		m.Variables = old.Variables
		m.Files = old.Files
		m.Pages = old.Pages
		m.PermissionSets = old.PermissionSets
		m.Site = old.Site
		m.Themes = old.Themes
		m.SitePermissionSets = old.Profiles
	}

	// unmarshal the current fields and join them
	current := struct {
		Profiles []string `json:"sitepermissionsets"`
	}{}
	err = json.Unmarshal(data, &current)
	if err != nil {
		return err
	} else {
		// just append
		m.SitePermissionSets = append(m.SitePermissionSets, current.Profiles...)
	}

	return nil
}
