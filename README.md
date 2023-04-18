# skuid CLI

skuid CLI is a command line interface for deploying and retrieving Skuid metadata between environments.

This repository's main branch currently represents the in-progress development of the next major release of the project.

To download binaries for the current stable release (0.5.0), see [this repository's releases page.](https://github.com/skuid/skuid-cli/releases)

To read documentation for the current stable release, [see the Skuid docs site.](https://docs.skuid.com/nlx/v2/en/skuid/cli/) 

## Local testing

To retrieve/deploy using a branch of skuid CLI, run the command in the root of this directory with go run main.go prepended. For example:
### Retrieve

To retrieve run ```go run main.go retrieve --host='site.pliny.webserver' -d /directory -u='user' -p='pass' -v```
To get more information about retrieve flags, use ```go run main.go retrieve --help```

### Deploy

To deploy run ```go run main.go deploy --host='site.pliny.webserver' -d /directory -u='user' -p='pass' -v```
To get more information about deploy flags, use ```go run main.go deploy --help```