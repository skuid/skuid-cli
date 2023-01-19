package pkg_test

import (
	"testing"

	"github.com/skuid/skuid-cli/pkg"
	"github.com/stretchr/testify/assert"
)

// TODO: uncomment
// func TestJsonUnmarshalMetadata(t *testing.T) {
// 	for _, tc := range []struct {
// 		description string
// 		given       string
// 		expected    []string
// 	}{
// 		{
// 			description: "profiles",
// 			given: `
// 			{
// 			 "host": "",
// 			 "port": "",
// 			 "url": "",
// 			 "type": "",
// 			 "metadata": {
// 				"apps": null,
// 				"authproviders": null,
// 				"componentpacks": null,
// 				"dataservices": null,
// 				"datasources": null,
// 				"designsystems": null,
// 				"variables": null,
// 				"files": null,
// 				"pages": null,
// 				"permissionsets": null,
// 				"sitepermissionsets": ["A", "B"],
// 				"site": null,
// 				"themes": null
// 			 },
// 			 "warnings": null
// 			}`,
// 			expected: []string{"A", "B"},
// 		},
// 		{
// 			description: "sitepermissionsets",
// 			given: `
// 			{
// 			 "host": "",
// 			 "port": "",
// 			 "url": "",
// 			 "type": "",
// 			 "metadata": {
// 				"apps": null,
// 				"authproviders": null,
// 				"componentpacks": null,
// 				"dataservices": null,
// 				"datasources": null,
// 				"designsystems": null,
// 				"variables": null,
// 				"files": null,
// 				"pages": null,
// 				"permissionsets": null,
// 				"profiles": ["A", "B"],
// 				"site": null,
// 				"themes": null
// 			 },
// 			 "warnings": null
// 			}`,
// 			expected: []string{"A", "B"},
// 		},
// 	} {
// 		t.Run(tc.description, func(t *testing.T) {
// 			plan := pkg.NlxPlan{}
// 			err := json.Unmarshal([]byte(tc.given), &plan)
// 			if err != nil {
// 				t.Log(err)
// 				t.FailNow()
// 			}
// 			assert.Equal(t, tc.expected, plan.Metadata.SitePermissionSets)
// 		})
// 	}
// }

func TestGetMetadataByString(t *testing.T) {

	badErr := pkg.GetFieldValueByNameError("bad")

	metadata := pkg.NlxMetadata{
		Apps:               []string{"apps"},
		AuthProviders:      []string{"authproviders"},
		ComponentPacks:     []string{"componentpacks"},
		DataServices:       []string{"dataservices"},
		DataSources:        []string{"datasources"},
		DesignSystems:      []string{"designsystems"},
		Variables:          []string{"variables"},
		Files:              []string{"files"},
		Pages:              []string{"pages"},
		PermissionSets:     []string{"permissionsets"},
		SitePermissionSets: []string{"sitepermissionsets"},
		Site:               []string{"site"},
		Themes:             []string{"themes"},
	}

	for _, tc := range []struct {
		description string
		given       string
		expected    []string
		expectedErr *error
	}{
		{
			description: "apps",
			given:       "apps",
			expected:    metadata.Apps,
		},
		{
			description: "authproviders",
			given:       "authproviders",
			expected:    metadata.AuthProviders,
		},
		{
			description: "componentpacks",
			given:       "componentpacks",
			expected:    metadata.ComponentPacks,
		},
		{
			description: "dataservices",
			given:       "dataservices",
			expected:    metadata.DataServices,
		},
		{
			description: "datasources",
			given:       "datasources",
			expected:    metadata.DataSources,
		},
		{
			description: "designsystems",
			given:       "designsystems",
			expected:    metadata.DesignSystems,
		},
		{
			description: "variables",
			given:       "variables",
			expected:    metadata.Variables,
		},
		{
			description: "files",
			given:       "files",
			expected:    metadata.Files,
		},
		{
			description: "pages",
			given:       "pages",
			expected:    metadata.Pages,
		},
		{
			description: "permissionsets",
			given:       "permissionsets",
			expected:    metadata.PermissionSets,
		},
		{
			description: "sitepermissionsets",
			given:       "sitepermissionsets",
			expected:    metadata.SitePermissionSets,
		},
		{
			description: "site",
			given:       "site",
			expected:    metadata.Site,
		},
		{
			description: "themes",
			given:       "themes",
			expected:    metadata.Themes,
		},
		{
			description: "bad",
			given:       "bad",
			expectedErr: &badErr,
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			for _, thing := range []string{} {
				x, err := metadata.GetFieldValueByName(thing)
				assert.Equal(t, []string{thing}, x)
				if err != nil {
					if tc.expectedErr != nil {
						assert.Equal(t, *tc.expectedErr, err)
					}
				}
			}
		})
	}
}
