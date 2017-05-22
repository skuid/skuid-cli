# skuid

`skuid` is a command line application for interacting with Skuid metadata, on both Salesforce Platform and Skuid Platform.
It provides ways to deploy and retrieve Skuid application metadata to and from either a Salesforce Platform org running the Skuid managed package or a Skuid Platform Site.

## Motiviation

Our previous offering [`skuid-grunt`](https://bitbucket.org/skuid/skuid-grunt) provided this functionality as a Grunt plugin.
It allowed users to configure Grunt tasks to perform pushes and pulls of Skuid Pages. This was great for projects already using NodeJS and Grunt
but not so great if you really didn't want to require those dependencies. `skuid` solves that problem by producing a self-contained
CLI to perform all the same operations.

## Examples (Salesforce Platform)

Pull all Skuid pages in the `Dashboard` module

```
$ skuid pull -m Dashboard
```

Push all Skuid pages in the `Dashboard` module

```
$ skuid push -m Dashboard
```

Create a Page Pack from all pages in the Dashboard module

```
$ skuid page-pack -m Dashboard -o src/staticresources/DashboardPages.resource
```

## Examples (Skuid Platform)

Retrieve all available Skuid Platform metadata from a given Site

```
$ skuid retrieve
```

Deploys all local Skuid Platform metadata to a Skuid Platform Site

```
$ skuid deploy
```

## Installation

If you have go installed on your system, you can simply `go get github.com/skuid/skuid`. If not, download the latest release from
[here](https://github.com/skuid/skuid/releases). You'll need to make sure `skuid` is in your `$PATH` and executable.

## Usage

```bash
Push and Pull Skuid pages to and from Skuid running on the Salesforce Platform

Usage:
  skuid [command]

Available Commands:
  page-pack   Retrieve a collection of Skuid Pages as a Page Pack.
  pull        Pull Skuid Pages from Salesforce into a local directory.
  push        Push Skuid Pages from a directory to Skuid.

Flags:
          --api-version string     API Version for Salesforce (default "39.0") / Skuid Platform (default "1")
          --client-id string       OAuth Client ID
          --client-secret string   OAuth Client Secret
      -d  --dir string             Input / output directory to use for a given command
      -h  --host string            Salesforce Login Url (default "https://login.salesforce.com") or Skuid Platform Site Url (e.g. "acme-us-trial.skuidsite.com")
      -m  --module string          Module name(s) to use for pull / push / page-pack commands
      -p  --password string        Password
      -u  --username string        Username


Use "skuid [command] --help" for more information about a command.
```

## Configuration

`skuid` uses Environment variables by default to provide credentials for interacting with the Skuid APIs under the hood.
The following Environment variables should be set to avoid having to enter your credentials with every command. You can drop these in your
`~/.bash_profile` or a `.env` file in your project directory.


### Salesforce Platform Example

```bash
export SKUID_UN={username}
export SKUID_PW={password + salesforce-security-token}
export SKUID_CLIENT_ID={connected-app-consumer-key}
export SKUID_CLIENT_SECRET={connected-app-consumer-secret}
```

### Salesforce Platform Example

```bash
export SKUID_UN={username}
export SKUID_PW={password}
export SKUID_HOST={host}
```

## Building from source

* To build from source for your machine
```
$ go build
```

* Building for a specific platform
```
$ GOOS=linux GOARCH=amd64 go build
# or
$ GOOS=linux GOARCH=amd64 make build #requires docker
```

## Future

* Support for deploy on Skuid Platform
* Support for watch behavior (e.g. skuid watch pages/*)
* Support for other Skuid objects on Salesforce Platform (Data Source, Theme, etc.)