# skuid

`skuid` is a command line application for interacting with Skuid metadata on both Salesforce Platform and Skuid Platform.
It provides ways to deploy and retrieve Skuid application metadata to and from either a Salesforce Platform org running the Skuid managed package or a Skuid Platform Site.

## Motiviation

Our previous offering [`skuid-grunt`](https://bitbucket.org/skuid/skuid-grunt) provided this functionality as a Grunt plugin.
It allowed users to configure Grunt tasks to perform pushes and pulls of Skuid Pages. This was great for projects already using NodeJS and Grunt,
but not so great if you really didn't want to require those dependencies. `skuid` solves that problem by producing a self-contained
CLI to perform all the same operations.

## Installation

If you have go installed on your system, you can simply `go get github.com/skuid/skuid`. If not, download the latest release from
[here](https://github.com/skuid/skuid/releases). You'll need to make sure `skuid` is in your `$PATH` and executable.

### Building from source

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

### Skuid Platform Example

```bash
export SKUID_UN={username}
export SKUID_PW={password}
export SKUID_HOST=https://humboldt-us-trial.skuidsite.com
```

## Usage

```bash
Usage:
  skuid [command]

Available Commands:
  deploy      Deploy local Skuid metadata to a Skuid Platform Site.
  page-pack   Retrieve a collection of Skuid Pages as a Page Pack.
  pull        Pull Skuid Pages from Salesforce into a local directory.
  push        Push Skuid Pages from a directory to Skuid.
  retrieve    Retrieve Skuid metadata from a Site into a local directory.

Flags:
      --api-version string     API Version
      --client-id string       OAuth Client ID
      --client-secret string   OAuth Client Secret
  -d, --dir string             Input/output directory.
      --host string            API Host base URL, e.g. my-site.skuidsite.com for Skuid Platform or my-domain.my.salesforce.com for Salesforce
  -m, --module string          Module name(s), separated by a comma.
  -p, --password string        Skuid Platform / Salesforce Password
  -u, --username string        Skuid Platform / Salesforce Username
  -v, --verbose                Display all possible logging info


Use "skuid [command] --help" for more information about a command.
```

### Examples (Salesforce Platform)

* Pull all Skuid pages in the `Dashboard` module

```
$ skuid pull -m Dashboard
```

* Push all Skuid pages in the `Dashboard` module

```
$ skuid push -m Dashboard
```

* Create a Page Pack from all pages in the Dashboard module

```
$ skuid page-pack -m Dashboard -o src/staticresources/DashboardPages.resource
```

### Examples (Skuid Platform)

* Retrieve all Skuid Platform metadata from a given Site, into the current directory

```
$ skuid retrieve
```

* Retrieve all Skuid Platform metadata from a given Site, into a specified directory

```
$ skuid retrieve -d sites/humboldt-us-trial
```

* Deploy all metadata in the current directory to a Skuid Platform Site

```
$ skuid deploy
```

* Deploy all metadata in a different directory to a Skuid Platform Site

```
$ skuid deploy -d path/to/directory
```



## Future

* Support for watch behavior (e.g. skuid watch pages/*)
* Support for other Skuid objects on Salesforce Platform (Data Source, Theme, etc.)