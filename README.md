# skuid CLI

skuid CLI is a command line interface for deploying and retrieving Skuid metadata between environments.

This repository's main branch currently represents the in-progress development of the next major release of the project.

To download binaries for the current stable release, see [this repository's releases page.](https://github.com/skuid/skuid-cli/releases)

To read documentation for the current stable release, [see the Skuid docs site.](https://docs.skuid.com/nlx/v2/en/skuid/cli/) 

## Go Version

The minimum version of Go supported is identified in this repositories [go.mod](go.mod#L4).

To change the Go version, the following must also be updated:

1. [dockerfile](compose/Dockerfile#L1)
1. [drone ci](.drone.yml#L20)
1. [makefile](Makefile#L6)

## Environment Variables

See [.env.template](.env.template) for a template to create a `.env` file from.

## Running the CLI

To retrieve/deploy using a branch of skuid CLI, run the command in the root of this directory with go run main.go prepended. For example:

### Retrieve

To retrieve run ```go run main.go retrieve --host='site.pliny.webserver' -d /directory -u='user' -p='pass' -v```
To get more information about retrieve flags, use ```go run main.go retrieve --help```

### Deploy

To deploy run ```go run main.go deploy --host='site.pliny.webserver' -d /directory -u='user' -p='pass' -v```
To get more information about deploy flags, use ```go run main.go deploy --help```

## Debugging the CLI

There are several configurations created for debug support in [Visual Studio Code](.vscode/launch.json):

1. `Retrieve w/ args` - CLI arguments passed on the command-line
2. `Retrieve w/ envvars` - Uses [.env](#environment-variables)
3. `Deploy w/ envvars` - Uses [.env](#environment-variables)
4. `Watch w/ envvars` - Uses [.env](#environment-variables)

## Testing the CLI

The [testify](https://github.com/stretchr/testify) toolkit is used to aid in writing tests, see [zip_test.go](./pkg/zip_test.go) for examples on usage.

### Generating/Updating Mocks

> [!NOTE]
> The CI/CD build process does not currently contain a step to re-build mocks prior to build/test.  In order to support automatically, the workflows & makefile need to be updated to execute `mockery`.  When generating/updating mocks, please ensure you manually run `mockery` prior to final commit to ensure you have generated the latest mocks based on your changes.

To automate the creation of mocks, the [mockery](https://github.com/vektra/mockery) tool is used.  While there are several tools in the eco-system to generate mocks automatically, `mockery` was chosen for the time being due to its tight integration with `testify`.  See [zip_test.go](./pkg/zip_test.go) for an example of using using mocks.

Where it makes sense, a mock may have a corresponding builder which is manually created to simplify creating mocks in tests.  See [FileUtilBuilder](./pkg/util/testutil/file_util_builder.go) for an example of a builder.

To generate new mocks or update existing mocks:

1. Follow the instructions to install [mockery](https://vektra.github.io/mockery/latest/installation/).
2. If updating a mock for an existing interface simply run `mockery` to re-generate all mocks.
3. If generating a new mock, update the `interfaces` list for the corresponding `package` (the package may need to be added) in [.mockery.yaml](.mockery.yaml) and run `mockery` to (re)generate all mocks.