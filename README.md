# skuid

`skuid` is a command line application for interacting with Skuid Pages on the Salesforce platform.
It provides ways to push and pull Skuid pages to and from your Salesforce org running the Skuid managed package.

## Motiviation

Our previous offering [`skuid-grunt`](https://bitbucket.org/skuid/skuid-grunt) provided this functionality as a Grunt plugin.
It allowed users to configure Grunt tasks to perform these pushes and pulls. This was great for projects already using NodeJS and Grunt
but not so great if you really didn't want to require those dependencies. `skuid` solves that problem by producing a self-contained
CLI to perform all the same operations.

## Examples

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

## Installation

If you have go installed on your system, you can simply `go get github.com/skuid/skuid`. If not, download the latest release from
[here](https://github.com/skuid/skuid/releases). You'll need to make sure `skuid` is in your `$PATH` and executable.

## Usage

```bash
Push and Pull Skuid pages to and from Skuid running on the Salesforce Platform

Usage:
  skuid [command]

Available Commands:
  page-pack   A brief description of your command
  pull        Pull Skuid Pages from Salesforce into a local directory.
  push        Push Skuid Pages from a directory to Skuid.

Flags:
      --api-version string     Salesforce API Version (default "36.0")
      --client-id string       Connected App Client ID
      --client-secret string   Connected App Client Secret
      --host string            Salesforce Login Url (default "https://login.salesforce.com")
      --password string        Salesforce Password
      --username string        Salesforce Username

Use "skuid [command] --help" for more information about a command.
```

## Configuration

`skuid` uses Environment variables by default to provide Salesforce credentials for interacting with the Salesforce API under the hood.
The following Environment variables should be set to avoid having to enter your credentials with every command. You can drop these in your
`~/.bash_profile` or a `.env` file in your project directory.

```bash
export SKUID_UN={salesforce-username}
export SKUID_PW={salesforce-password + salesforce-security-token}
export SKUID_CLIENT_ID={connected-app-client-id}
export SKUID_CLIENT_SECRET={connected-app-client-secret}
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