# skuid CLI

`skuid` is a command line application for retrieving and deploying Skuid objects (data and metadata) on both Skuid Platform sites and Salesforce orgs running the Skuid managed package.

While Skuid is a cloud user experience platform, it can be helpful download an entire site's worth of pages to make small adjustments on multiple pages, to move them from a sandbox site to a production site, or to store them in a version control system (VCS). And while it's possible to save page XML and create page packs, moving entire swathes of metadata from site to site can prove challenging.

Enter the ``skuid`` command-line application. Using ``skuid`` you can easily pull—or download—your Skuid pages and push—upload—your Skuid metadata from one site to another using only a few commands.

* [Prerequisites](#prerequisites)
* [Installation](#installation)
	* [Building from source](#building-from-source)
* [Configuration](#configuration)
* [Usage](#usage)
  * [Commands](#commands)
  * [Command Flags](#command-flags)
  * [Examples](#examples)
    * [Skuid Platform](#skuid-platform)
    * [Salesforce Platform](#salesforce-platform)
* [Skuid Object Structure](#skuid-object-structure)
  * [What Is Not Retrieved by skuid](#what-is-not-retrieved-by-skuid)
* [Use Cases](#use-cases)
	* [Version control](#version-control)
	* [Sandboxes](#sandboxes)
* [Troubleshooting](#troubleshooting)
* [skuid vs skuid-grunt](#skuid-vs-skuid-grunt)
* [Future](#future)

## Prerequisites

- A UNIX-based CLI, meaning you'll need to be working on a macOS or Linux machine.
- Some basic knowledge of the command line is required. You'll need to be able to navigate your file system with `cd` and enter the `skuid` command. 
- If pulling and pushing pages with Skuid on Salesforce...
  - You must [enable My Domain](https://help.salesforce.com/articleView?id=domain_name_overview.htm&type=5) for your Salesforce org
  - You must configure [Salesforce connected app](https://help.salesforce.com/articleView?id=connected_app_overview.htm&type=5) and have both the the _consumer key_ and the _consumer secret_.
  
## Installation

To manually install the application, follow these steps:

1. Download [the latest release of the `skuid` CLI application binary.](https://github.com/skuid/skuid/releases)
1. Rename the downloaded application binary file to ``skuid``.
1. Move the application to a folder in your `$PATH` variable, like `/usr/local/bin`, or add the application's folder to the PATH variable.
1. Give the application executable permissions: `chmod +x skuid`
1. Verify that you can run the application: ``skuid --help``.

You may also copy and paste the following commands in your terminal:

```
curl -L https://github.com/skuid/skuid/releases/download/0.2.0/skuid_darwin_amd64 -o skuid
chmod +x skuid
mv skuid /usr/local/bin/
```

*Note*: If you are on a Linux machine, download the appropriate version of the application with: 
```bash
curl -L https://github.com/skuid/skuid/releases/download/0.2.0/skuid_linux_amd64 -o skuid
```

*Note:* Use the [the Go programming language?](https://golang.org/doc/install) If so, you can also install ``skuid`` by running `go get github.com/skuid/skuid`.

### Building from source

To build the application from the source, you'll first need to clone the repository or download the source. You'll also need [Go installed on your machine](https://golang.org/doc/install).

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

`skuid` uses environment variables by default to provide credentials for interacting with the Skuid APIs under the hood. While you can set your username, password, host, and connected app settings with flags, consider setting the following environment variables to avoid entering your credentials with every command. 

Enter the appropriate information in the format listed below, listing your own username, password, etc. immediately following the `=` equals sign. You can drop these in your `~/.bash_profile`, `~/.zshrc`, or a `.env` file in your project directory.

Which environment variables you'll need to set varies depending on which platform you'll be connecting to:

**Skuid Platform**

```bash
export SKUID_UN={username}
export SKUID_PW={password}
export SKUID_HOST={https://example.skuidsite.com}
```

**Skuid on Salesforce**

```bash
export SKUID_UN={username}
export SKUID_PW={password + salesforce-security-token}
export SKUID_HOST={https://my-domain.my.salesforce.com}
export SKUID_CLIENT_ID={connected-app-consumer-key}
export SKUID_CLIENT_SECRET={connected-app-consumer-secret}
```

Note that your `SKUID_PW` in this case must be a user's Salesforce password directly connected to [the user's Salesforce security token](https://help.salesforce.com/articleView?id=user_security_token.htm&type=5). 

So with `AMostExcellentPassword` as a password and `aBc12dEF34gh56ij7k` as a security token, the `SKUID_PW` would be exported with:

```bash
export SKUID_PW=AMostExcellentPasswordaBc12dEF34gh56ij7k
```

## Usage

`skuid` is used to do two things:
1. **Retrieve/pull** Skuid data and metadata from a platform hosting Skuid and **store** it on the local machine
1. **Deploy/push** Skuid data and metadata from the local machine to a platform hosting Skuid

All of this is done using the following syntax:

```bash
  skuid [command] [flags]
```

### Commands

The commands you'll be using to accomplish this depend on your platform of choice:
  - When using **Skuid Platform**, you can use these commands:
    - `retrieve` will retrieve data from the Skuid Platform site.
    - `deploy` will send data to the Skuid Platform site.
  
  - When using **Skuid on Salesforce**, you can use these commands:
    - `pull` will retrieve Skuid pages from the Salesforce org.
      - Can be used with the `--module` flag to pull one or more specified modules.
    - `push` will send Skuid pages to the Salesforce org.
      - Can be used with the `--file` flag to push a specific Skuid page.
    - `page-pack` will retrieve Skuid pages from the Salesforce org as a [page pack](https://docs.skuid.com/latest/en/skuid/pages/import-export-page-packs-modules.html#what-is-a-page-pack).
       - **Requires** the `--output` flag.
       - Can be used with the `--module` flag to pull one or more specified modules.

### Command Flags

- **Authentication and Platform**: These flags can be used when authenticating to *either platform* in lieu of [exporting environment variables](#configuration).
  - `--host`:  (string)  The Skuid host platform's base URL, e.g. https://example.skuidsite.com for Skuid Platform or https://my-domain.my.salesforce.com for Salesforce
  - `--password`: (string) Skuid Platform / Salesforce Password
    - Abbreviated form: `-p`
  - `--username`: (string) Skuid Platform / Salesforce Username
    - Abbreviated form: `-u`
  - `--client-id` (string): **Skuid on Salesforce only.** The consumer ID for the Salesforce connected app.
  - `--client-secret` (string): **Skuid on Salesforce only.** The consumer Secret for the Salesforce connected app.
  - `--api-version`: (number) **Skuid Platform only.** Select which version of the deployment API to use. Only version `1` is active at this time.
- **Data management**: 
  - `--dir`: (string) The input/output directory where files are retrieved and stored to *or* deployed from.
    - Abbreviated form: `-d`
  - `--module`: (string) **Skuid on Salesforce only.** One or more Skuid page modules, separated by commas, to deploy or retrieve.
    - Abbreviated form: `-m`
- **Debugging**:
  - `--verbose`: When used, `skuid` will display all possible logging information.
    - Abbreviated form: `-v`
- **Command-specific**:
  - `--file`: (string) Used with `skuid push` to push a specific Skuid page to a Salesforce org. **Must point to a page's `.json` file.** 
    - Abbreviated form: `-f` 
  - `--output`: (string) Used with `skuid page-pack` to set the filename of the created page pack.
    - Abbreviated form: `-o` 

###  Examples 

#### Skuid Platform

* Retrieve all Skuid data from a Skuid Platform site and store in the current directory

  ```
  $ skuid retrieve
  ```

* Retrieve all Skuid data from a Skuid Platform site and store in a specified directory

  ```
  $ skuid retrieve -d sites/humboldt-us-trial
  ```

* Deploy all data in the current directory to a Skuid Platform site

  ```
  $ skuid deploy
  ```

* Deploy all data in a different directory to a Skuid Platform site

  ```
  $ skuid deploy -d path/to/directory
  ```

#### Salesforce Platform

* Pull all Skuid pages in the Salesforce org

  ```
  $ skuid pull
  ```

* Pull all Skuid pages in the `Dashboard` module

  ```
  $ skuid pull -m Dashboard
  ```

* Push all Skuid pages in the `Dashboard` module

  ```
  $ skuid push -m Dashboard
  ```

* Create a page pack consisting of all pages in the Dashboard module

  ```
  $ skuid page-pack -m Dashboard -o src/staticresources/DashboardPages.resource
  ```

## Skuid Object Structure

When pulling from the Salesforce platform, all pages are stored within a `skuidpages` directory in the current working directory. Each page consists of two files:
  - A `.json` file, consisting of the page's metadata
  - The page's `.xml` file

When retrieving from Skuid Platform sites, the following is downloaded:

- All Skuid pages in the `pages` directory
- All [apps, and the routes within them](https://docs.skuid.com/latest/en/skuid-platform/deploying-apps-in-skuid-platform.html), in the `apps` directory
- All [authentication providers](https://docs.skuid.com/latest/en/data/) in the `authproviders` directory
- All [data sources](https://docs.skuid.com/latest/en/data/) in the `datasources` directory
- All [profiles](https://docs.skuid.com/latest/en/skuid-platform/user-and-permission-management.html#profiles) in the `profiles` directory

### What Is Not Retrieved by skuid

- **Created by** and **Modified by** metadata for pages
  - When deploying Skuid pages to a site, the created/modified by user and date will match **the identity of whoever is running the deployment**, as well as **the time and date of the deployment.**
- **Skuid Platform:**
  - Authentication provider credentials
    - **You must re-enter any client ID and client secret pairs on all Skuid authentication providers in the target site**, even if those authentication providers already existed.
  - Users and user data for Skuid Platform
    - While user profiles are transferred, individual user accounts and their information are not. Users must be manually recreated—or [provisioned through single sign-on](https://docs.skuid.com/latest/en/skuid/single-sign-on/#user-provisioning-within-skuid-platform)—for at least the first deployment.
  - Site Settings (offline mode, single sign-on configurations, security and frame embedding options, etc.)

## Use Cases

### Version control

Retrieving Skuid data objects for local storage allows for the use of version control systems, such as `git`. 

For example, you could create a directory for your Skuid site and intiate a `git` repository:
```bash
mkdir my-skuid-site
cd my-skuid-site
git init
```

And, after exporting the proper authentication credentials, you could `retrieve` and commit your Skuid data within that `git` repo like so:
```bash
skuid retrieve
git add -A
git commit -m 'Initial commit of Skuid site'
```

With a workflow like this, it's easier to capture snapshots of Skuid sites at certain points in time, allowing both for easy rollbacks should issues arise, as well as developer collaboration on these Skuid objects. For a more in-depth tuotrial, see [Version Control with Salesforce - A Primer](https://github.com/skuid/sfdc-vcs-tutorial).

### Sandboxes

`skuid` is particularly useful for moving Skuid data from a sandbox—or testing space—to production for end-users. Whether that be Skuid on Salesforce—which takes [some additional deployment steps](https://docs.skuid.com/latest/en/skuid/deploy/salesforce/org-to-org.html)—or the nearly complete transferral possible for Skuid Platform sites. 

The `-u`, `-p` and `--host` flags can be used to temporarily set different authentication and platform settings. 

For example, consider a workflow like below:

```bash
# Set the authentication credentials for the sandbox site, assuming they are not set already
export SKUID_UN={sandbox-username}
export SKUID_PW={sandbox-password}
export SKUID_HOST={https://example-sandbox.skuidsite.com}
# Retreieve sandbox data
skuid retrieve
# Deploy sandbox data to production by temporarily using different credentials
skuid deploy -u production-username -p production-password  --host https://production-sandbox.skuidsite.com}
```

You could even consider automating this process using shell scripts to suit your needs. Experiment to find the workflow that best suits your lifecycle management processes.

## Troubleshooting

By default `skuid` tries to only show basic information so as not to clutter the terminal. If you're seeing errors, a good first step is try your command again with the `-v` or `--verbose` flag to log more information. 

Some errors you may encounter include:

- `unexpected end of JSON input`
  - This error can appear when attempting to deploy to a Skuid Platform site. This may mean that your SKUID_UN and SKUID_PW variables are not set.
- `Error deploying metadata: Post https://example.skuidsite.com/api/v1/metadata/deploy: EOF`
  - This may mean your Skuid credentials are not correct within your environment variables. Verify both your username and password.
- `Post example.my.salesforce.com/services/oauth2/token: unsupported protocol scheme ""`
  - Your ``SKUID_HOST`` or ``--host`` flag value may not have the `https` protocol in the proper place.

    **Correct**

    ```
    export SKUID_HOST=https://example.my.salesforce.com
    ```

    **Incorrect**
    ```
    export SKUID_HOST=example.my.salesforce.com
    ```
- `Get /services/apexrest/skuid/api/v1/pages?module=: unsupported protocol scheme ""`
  - You may not have your `SKUID_HOST` set appropriately. If you're pushing or deploying to a different site temporarily, consider using the `--host` flag to temporarily point to another host.
  - If you are positive you're pointing at the correct host, your `SKUID_PW` variable may not be set correctly. Ensure that both your Salesforce password _and_ your security token are set—written together with no characters between them—for this variable.
- `Error retrieving metadata:  Error making HTTP request%!(EXTRA string=401 Unauthorized)`
  - Your `SKUID_UN` or `SKUID_PW` may not be set correctly. Verify your user credentials.
- `Error deploying metadata: Error making HTTP request%!(EXTRA string=400 Bad Request)` or `Error deploying metadata: Error making HTTP request%!(EXTRA string=500 Internal Server Error)`
  - These errors indicate there may be an issue with the retrieval[your Skuid Platform apps.](https://docs.skuid.com/latest/en/skuid-platform/deploying-apps-in-skuid-platform.html#apps)
    - Ensure that all apps you have retrieved are properly configured, with correct routes and corresponding pages for those routes.
    - Verify there are no issues with [user profile access](https://docs.skuid.com/latest/en/skuid-platform/user-and-permission-management.html#profiles) to the apps you are attempting to deploy.
    - If all the options appear correct, **delete the `apps` folder** on your local machine and attempt `skuid deploy` again. Recreate your apps manually on the destination Platform site.

## skuid vs skuid-grunt

[`skuid-grunt`](https://bitbucket.org/skuid/skuid-grunt) previously provided this push/pull functionality as a Grunt plugin. This was great for projects already using NodeJS and Grunt, but not so great if you didn't want to require those dependencies. `skuid` solves that problem by producing a self-contained
CLI to perform all the same operations.

## Future

* Support for watch behavior (e.g. skuid watch pages/*)
* Support for other Skuid objects on Salesforce Platform (Data Source, Theme, etc.)