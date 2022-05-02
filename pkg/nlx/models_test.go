package nlx_test

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
// 			plan := nlx.NlxPlan{}
// 			err := json.Unmarshal([]byte(tc.given), &plan)
// 			if err != nil {
// 				t.Log(err)
// 				t.FailNow()
// 			}
// 			assert.Equal(t, tc.expected, plan.Metadata.SitePermissionSets)
// 		})
// 	}
// }
