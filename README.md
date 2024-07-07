# skuid CLI

skuid CLI is a command line interface for deploying and retrieving Skuid metadata between environments.

This repository's main branch currently represents the in-progress development of the next major release of the project.

To download binaries for the current stable release, see [this repository's releases page.](https://github.com/skuid/skuid-cli/releases)

To read documentation for the current stable release, [see the Skuid docs site.](https://docs.skuid.com/nlx/v2/en/skuid/cli/) 

## Go Version

The minimum version of Go supported is identified in this repositories [go.mod](go.mod#L4).

To change the Go version, the following must also be updated:

1. [github workflow](.github/workflows/github-actions-release.yml#L24)
2. [dockerfile](compose/Dockerfile#L1)
3. [drone ci](.drone.yml#L20)
4. [makefile](Makefile#L6)

## Local testing

To retrieve/deploy using a branch of skuid CLI, run the command in the root of this directory with go run main.go prepended. For example:

### Retrieve

To retrieve run ```go run main.go retrieve --host='site.pliny.webserver' -d /directory -u='user' -p='pass' -v```
To get more information about retrieve flags, use ```go run main.go retrieve --help```

### Deploy

To deploy run ```go run main.go deploy --host='site.pliny.webserver' -d /directory -u='user' -p='pass' -v```
To get more information about deploy flags, use ```go run main.go deploy --help```

## Local debugging

There are several configurations created for debug support in [Visual Studio Code](.vscode/launch.json):

1. `Retrieve w/ args` - CLI arguments passed on the command-line
2. `Retrieve w/ envvars` - Uses [.env](#environment-variables)
3. `Deploy w/ envvars` - Uses [.env](#environment-variables)
4. `Watch w/ envvars` - Uses [.env](#environment-variables)

## Environment Variables

See [.env.template](.env.template) for a template to create a `.env` file from.