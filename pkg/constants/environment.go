package constants

const (
	ENV_PREFIX = "SKUID"

	ENV_PLINY_HOST             = "PLINY_HOST"                // legacy for SKUID_HOST
	ENV_SKUID_PW               = "SKUID_PW"                  // legacy for SKUID_PASSWORD
	ENV_SKUID_UN               = "SKUID_UN"                  // legacy for SKUID_USERNAME
	ENV_SKUID_APP_NAME         = "SKUID_APP_NAME"            // legacy for SKUID_APP
	ENV_SKUID_DEFAULT_FOLDER   = "SKUID_DEFAULT_FOLDER"      // legacy for SKUID_DIR
	ENV_SKUID_LOGGING_LOCATION = "SKUID_LOGGING_DIRECTORY"   // legacy for SKUID_LOG_DIRECTORY
	ENV_SKUID_RETRIEVE_SINCE   = "SKUID_RETRIEVE_SINCE_DATE" // legacy for SKUID_SINCE
	ENV_SKUID_FIELD_LOGGING    = "SKUID_FIELD_LOGGING"       // legacy for SKUID_DIAGNOSTIC
	ENV_SKUID_IGNORE_SKUIDDB   = "SKUID_IGNORE_SKUIDDB"      // legacy for SKUID_IGNORE_SKUID_DB

	TEST_ENVIRONMENT_FILE = ".testenv"
)
