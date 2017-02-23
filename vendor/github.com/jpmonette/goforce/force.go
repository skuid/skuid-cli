package force

import (
	"encoding/json"
)

type ForceAPI struct {
	Connection *Connection
}

type OAuthResponse struct {
	Id          string `json:"id"`
	TokenType   string `json:"token_type"`
	InstanceUrl string `json:"instance_url"`
	Signature   string `json:"signature"`
	AccessToken string `json:"access_token"`
}

// Login is logging-in the User to the Salesforce organization
func Login(consumerKey string, consumerSecret string, instanceUrl string, username string, password string, apiVersion string) (api *ForceAPI, err error) {

	conn := Connection{
		ConsumerKey:    consumerKey,
		ConsumerSecret: consumerSecret,
		InstanceUrl:    instanceUrl,
		Username:       username,
		Password:       password,
		APIVersion:     apiVersion,
	}

	err = conn.Refresh()

	if err != nil {
		return api, err
	}

	api = &ForceAPI{
		Connection: &conn,
	}

	return api, err
}

// Limits is getting the current maximum and remaining limits
// of the organization.
func (api *ForceAPI) Limits() (data LimitsResponse, err error) {

	result, err := api.Connection.Get("/limits/")

	if err != nil {
		return LimitsResponse{}, err
	}

	err = json.Unmarshal(result, &data)

	if err != nil {
		return LimitsResponse{}, err
	}

	return data, nil
}

type LimitsResponse struct {
	ConcurrentAsyncGetReportInstances Limit
	ConcurrentSyncReportRuns          Limit
	DailyApiRequests                  Limit
	DailyAsyncApexExecutions          Limit
	DailyBulkApiRequests              Limit
	DailyGenericStreamingApiEvents    Limit
	DailyStreamingApiEvents           Limit
	DailyWorkflowEmails               Limit
	DataStorageMB                     Limit
	FileStorageMB                     Limit
	HourlyAsyncReportRuns             Limit
	HourlyDashboardRefreshes          Limit
	HourlyDashboardResults            Limit
	HourlyDashboardStatuses           Limit
	HourlySyncReportRuns              Limit
	HourlyTimeBasedWorkflow           Limit
	MassEmail                         Limit
	SingleEmail                       Limit
	StreamingApiConcurrentClients     Limit
}

type Limit struct {
	Remaining int
	Max       int
}
