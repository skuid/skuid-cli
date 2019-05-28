# skuid CLI

`skuid` is a command line interface (CLI) for retrieving and deploying Skuid objects (data and metadata) on both Skuid Platform sites and Salesforce orgs running the Skuid managed package.

While Skuid is a cloud user experience platform, it can be helpful to:

- download an entire site's worth of pages to make small adjustments locally
- move Skuid configurations from a sandbox site to a production site
- store Skuid configurations in a version control system (VCS)

While it's possible to save page XML and create page packs, moving entire apps' worth of Skuid objects from site to site can prove challenging.

Enter the `skuid` CLI. Using `skuid`, you can easily pull—download—Skuid pages and push—upload—Skuid metadata from one site to another using only a few commands.

_Note_: `skuid` CLI is open source, and we accept pull requests! If there is a feature you'd like to see, feel free to contribute.

## Prerequisites

- Some basic knowledge of the command line is required, including how to open and use a command line program such as Terminal or Powershell.
  - You should feel comfortable interacting with the file system with commands like `cd`, `mv`, and `rm`.
  - You should feel comfortable using the `skuid` command with the necessary environment variables and flags.
- If pulling and pushing pages with Skuid on Salesforce...
  - [Enable My Domain](https://help.salesforce.com/articleView?id=domain_name_overview.htm&type=5) for your Salesforce org
  - Configure a [Salesforce connected app](https://help.salesforce.com/articleView?id=connected_app_overview.htm&type=5) and have both the _consumer key_ and the _consumer secret_ available

## Installation

### macOS and Linux

If you have `curl`, `grep` and `awk` in your environment, you can quickly install the application via the command line:

```bash
# macOS:
curl -Lo skuid $(curl -sL https://api.github.com/repos/skuid/skuid-cli/releases/latest | grep "browser_download_url.*darwin" | awk -F '"' '{print $4}')

# Linux:
# curl -Lo skuid $(curl -sL https://api.github.com/repos/skuid/skuid-cli/releases/latest | grep "browser_download_url.*linux" | awk -F '"' '{print $4}')

# Give the skuid application the permissions it needs to run
chmod +x skuid

# Move the skuid application to a folder where it can be used easily, 
# for example, a directory in your $PATH.
# Enter your computer account password if prompted
sudo mv skuid /usr/local/bin/skuid
```

To manually install the application, follow these steps:

1. Download [the latest release of the `skuid` application binary.](https://github.com/skuid/skuid-cli/releases)
1. Navigate to the directory containing the `skuid` binary file in a terminal application.

   ```bash
   cd /path/to/the/downloaded/binary
   ```

1. Rename the downloaded application binary file to `skuid`:

   ```bash
   # macOS:
   mv skuid_darwin_amd64 skuid
   # Linux:
   # mv skuid_linux_amd64 skuid
   ```

1. Give the application executable permissions:

   ```bash
   chmod +x skuid
   ```

1. Move the application to a folder in your `$PATH` variable, like `/usr/local/bin`, or add the application's folder to the PATH variable:

   ```bash
   mv skuid /usr/local/bin/skuid
   # or add the below to your shell profile
   export PATH=$PATH:/path/to/a-folder
   ```

   ```eval_rst
   .. note:: If you choose to update your shell profile you'll need to `source` your shell profile or restart your session for those changes to take effect.
   ```

1. Verify that you can run the application:

   ```bash
   skuid --help
   ```

### Windows

1. Download the [latest Windows executable release](https://github.com/skuid/skuid-cli/releases) in your web browser.
1. (_Optional_) Move the executable to a more permanent location, such as `C:\Program Files\Skuid\`.
1. (_Optional_) Set an alias to easily access the executable.

   In Windows Powershell, use Set-Alias. For information about saving aliases in Powershell, see [Microsoft documentation](https://docs.microsoft.com/en-us/powershell/module/microsoft.powershell.utility/set-alias?view=powershell-6)

   ```bash
   Set-Alias skuid C:\Path\To\skuid_windows_amd64.exe
   ```

   In the Windows Command Prompt, use doskey. For information about saving doskey aliases for future sessions, see [Microsoft documentation](https://technet.microsoft.com/en-us/library/ff382652.aspx).

   ```bash
   doskey skuid=C:\Path\To\skuid_windows_amd64.exe
   ```

1. Invoke the executable by typing its name or alias and pressing enter: `skuid --help`.

### Using go

Use the [the Go programming language?](https://golang.org/doc/install) If so, you can also install `skuid` by running `go get github.com/skuid/skuid-cli`.

#### Building from source

To build the application from the source, first clone the repository or download the source. You also need [Go installed on your machine](https://golang.org/doc/install).

- To build from source for your machine:

  ```bash
  go build
  ```

- Building for a specific platform:

  ```bash
  GOOS=linux GOARCH=amd64 go build
  # or
  GOOS=linux GOARCH=amd64 make build #requires docker
  ```

### Upgrading skuid CLI

To upgrade `skuid`, first remove the previous version's binary:

```bash
# Use the rm command with the appropriate path to the skuid binary
# For example, if the binary is installed in /usr/local/bin:
rm /usr/local/bin/skuid
```

After removing the previous version, repeat [the installation steps listed above.](#installation)

## Configuration

By default, `skuid` uses environment variables to provide credentials for interacting with Skuid APIs under the hood. While you can set username, password, host, and connected app values with flags, consider setting the following environment variables to avoid entering credentials with every command.

Which environment variables you'll need to set depends on which platform you'll be connecting to.

### Skuid Platform

#### macOS and Linux

Enter the appropriate information in the format listed below, listing your own username, password, etc., immediately following the `=` equals sign. You can drop these in your `~/.bash_profile`, `~/.zshrc`, or into a `.env` file in the project directory.

```bash
export SKUID_UN='username'
export SKUID_PW='password'
export SKUID_HOST='https://example.skuidsite.com'
```

#### Windows

How you set your environment variables differs depending on your command line program of choice:

**Powershell**

```bash
Set-Item Env:SKUID_UN 'username'
Set-Item Env:SKUID_PW 'password'
Set-Item Env:SKUID_HOST 'https://example.skuidsite.com'
```
**Command Prompt**

```bash
Set SKUID_UN=username
Set SKUID_PW=password
Set SKUID_HOST=https://example.skuidsite.com
```

### Skuid on Salesforce

Note that the `SKUID_PW` in this case must be a user's Salesforce password directly connected to [the user's Salesforce security token](https://help.salesforce.com/articleView?id=user_security_token.htm&type=5).

So with `AMostExcellentPassword` as a password and `aBc12dEF34gh56ij7k` as a security token, the `SKUID_PW` would be set to:

`AMostExcellentPasswordaBc12dEF34gh56ij7k`

#### macOS and Linux

Enter the appropriate information in the format listed below, listing your own username, password, etc., immediately following the `=` equals sign. You can drop these in your `~/.bash_profile`, `~/.zshrc`, or into a `.env` file in the project directory.


```bash
export SKUID_UN='username'
export SKUID_PW='password + salesforce-security-token'
export SKUID_HOST='https://my-domain.my.salesforce.com'
export SKUID_CLIENT_ID='connected-app-consumer-key'
export SKUID_CLIENT_SECRET='connected-app-consumer-secret'
```

#### Windows

How you set your environment variables differs depending on your command line program of choice:

**Powershell**

```bash
Set-Item Env:SKUID_UN 'username'
Set-Item Env:SKUID_PW 'password + salesforce-security-token'
Set-Item Env:SKUID_HOST 'https://my-domain.my.salesforce.com'
Set-Item Env:SKUID_CLIENT_ID 'connected-app-consumer-key'
Set-Item Env:SKUID_CLIENT_SECRET 'connected-app-consumer-secret'
```

**Command Prompt**

```bash
Set SKUID_UN=username
Set SKUID_PW=password + salesforce-security-token
Set SKUID_HOST=https://my-domain.my.salesforce.com
Set SKUID_CLIENT_ID=connected-app-consumer-key
Set SKUID_CLIENT_SECRET=connected-app-consumer-secret
```

## Usage

`skuid` is used to do two things:

1. **Retrieve/pull** Skuid data and metadata _from_ a platform hosting Skuid and **store** it on the local machine
1. **Deploy/push** Skuid data and metadata _from_ the local machine _to_ a platform hosting Skuid

All of this is done using the following syntax:

```bash
  skuid [command] [flags]
```

**Warning**: It is a known issue that **the Windows executable is currently unable to deploy to Skuid Platform sites.**

### Commands

The commands used to accomplish this depend on the platform of choice:

  - When using **Skuid Platform**:
    - `retrieve` retrieves data from the Skuid Platform site.
    - `deploy` sends data to the Skuid Platform site.
  - When using **Skuid on Salesforce**:
    - `pull` retrieves Skuid pages from the Salesforce org.
      - Can be used with the `--module` flag to pull one or more specified modules.
    - `push` sends all Skuid pages within current directory to the Salesforce org.
      - Can be used with the `--file` flag to push a specific Skuid page.
      - Can be used with the `--module` flag to specify which module the page(s) should be added to. If no module is specified, then the page(s) will not be added to any module.
    - `page-pack` retrieves Skuid pages from the Salesforce org as a [page pack](https://docs.skuid.com/latest/en/skuid/pages/page-packs.html).
       - **Requires** the `--output` flag.
       - Can be used with the `--module` flag to pull one or more specified modules.

### Command Flags

- **Authentication and Platform**: These flags can be used when authenticating to _either platform_ in lieu of [exporting environment variables](#configuration).
  - `--host`:  (string)  The Skuid host platform's base URL, e.g. https://example.skuidsite.com for Skuid Platform or https://my-domain.my.salesforce.com for Salesforce
  - `--password`: (string) Skuid Platform / Salesforce Password
    - Abbreviated form: `-p`
  - `--username`: (string) Skuid Platform / Salesforce Username
    - Abbreviated form: `-u`
  - `--client-id` (string): **Skuid on Salesforce only.** The consumer ID for the Salesforce connected app.
  - `--client-secret` (string): **Skuid on Salesforce only.** The consumer secret for the Salesforce connected app.
  - `--dataServiceProxy` (string): The IP or URI through which traffic should be routed to reach a Skuid site's data services.  May be necessary for cases where skuid CLI is executed from a machine on an internal network and VPN rules require proxy use.
  - `--metadataServiceProxy` (string): The IP or URI through which traffic should be routed to reach a Skuid site's metadata service.  May be necessary for cases where skuid CLI is executed from a machine on an internal network and VPN rules require proxy use.
- **Data management**:
  - `--dir`: (string) The input/output directory where files are retrieved and stored to _or_ deployed from.
    - Abbreviated form: `-d`
  - `--module`: (string) **Skuid on Salesforce only.** One or more Skuid page modules, separated by commas, to deploy or retrieve.
    - Abbreviated form: `-m`
  - `--no-module`: (string) **Skuid on Salesforce only.**  Indicates that `skuid` should `pull` all pages that are **not included in a module**. May be used in conjunction with the `--module` flag to also pull pages from modules during the same `pull` action.

    Does not apply to the `push` command.

  - `--page`: (string) **Skuid on Salesforce only.** One or more Skuid pages, listed by page name and separated by commas, to retrieve to the local file system.
    - Abbreviated form: `-n`
- **Debugging**:
  - `--verbose`: When used, `skuid` displays all possible logging information.
    - Abbreviated form: `-v`
- **Command-specific**:
  - `--file`: (string) Used with `skuid push` to push a specific Skuid page to a Salesforce org. **Must point to a page's `.json` file, not its `.xml` file**. Even though edits are made to a page's `.xml` file, this flag will only work if it points to the `.json` file.
    - Abbreviated form: `-f`
  - `--output`: (string) Used with `skuid page-pack` to set the filename of the created page pack.
    - Abbreviated form: `-o`

  It is also possible to use the `--api-version` flag to select which version of the deployment API to use, however it does not have much use as only version `1` is active at this time.

### Examples

#### Skuid Platform

- Retrieve all Skuid data from a Skuid Platform site and store in the current directory:

  ```bash
  skuid retrieve
  ```

- Retrieve all Skuid data from a Skuid Platform site and store in a specified directory:

  ```bash
  skuid retrieve -d sites/humboldt-us-trial
  ```

- Deploy all data in the current directory to a Skuid Platform site:

  ```bash
  skuid deploy
  ```

- Deploy all data in a different directory to a Skuid Platform site:

  ```bash
  skuid deploy -d path/to/directory
  ```

#### Salesforce Platform

- Pull all Skuid pages in the Salesforce org:

  ```bash
  skuid pull
  ```

- Pull Skuid pages that **do not have a module** specified:

  ```bash
  skuid pull --no-module
  ```

- Pull all Skuid pages in the `Dashboard` module:

  ```bash
  skuid pull -m Dashboard
  ```

- Pull any Skuid page named `example` that exists within the `Dashboard` module _as well as_ any page named `example` that does not have a module specified:

  ```bash
  skuid pull -n example -m Dashboard --no-module
  ```

- Pull all pages in the `Dashboard` module and create a page pack called `DashboardPages`:

  ```bash
  skuid page-pack -m Dashboard -o src/staticresources/DashboardPages.resource
  ```

- Push all Skuid pages in the `Dashboard` module within the current directory:

  ```bash
  skuid push -m Dashboard
  ```

- Push all Skuid pages in `skuidpages` directory

  ```bash
  skuid push -d skuidpages
  ```

- Push only the `AccountTab` Skuid page in the `skuidpages` directory:

  ```bash
  skuid push -f skuidpages/AccountTab.json
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
  - When deploying Skuid pages to a site, the created/modified by user and date matches **the identity of whoever is running the deployment**, as well as **the time and date of the deployment.**
- **Skuid Platform:**
  - Authentication provider credentials
    - **You must re-enter any client ID and client secret pairs on all Skuid authentication providers in the target site**, even if those authentication providers already existed.
  - Users and user data for Skuid Platform
    - While user profiles are transferred, individual user accounts and their information are not. Users must be manually re-created—or [provisioned through single sign-on](https://docs.skuid.com/latest/en/skuid/single-sign-on/#user-provisioning-within-skuid-platform)—for at least the first deployment.
  - Site Settings (offline mode, single sign-on configurations, security and frame embedding options, etc.)

## Use Cases

### Version control

Retrieving Skuid data objects for local storage allows for the use of version control systems, such as `git`.

For example, you could create a directory for your Skuid site and initiate a `git` repository:

```bash
mkdir my-skuid-site
cd my-skuid-site
git init
```

After exporting the proper Skuid authentication credentials, you could `retrieve` and commit Skuid data within that `git` repo:

```bash
skuid retrieve
git add -A
git commit -m 'Initial commit of Skuid site'
```

With a workflow like this, it's easier to capture snapshots of Skuid sites at certain points in time, allowing for both easy rollbacks (should issues arise) as well as developer collaboration on these Skuid objects. For a more in-depth tutorial, see [Version Control with Salesforce - A Primer](https://github.com/skuid/sfdc-vcs-tutorial).

### Sandboxes

`skuid` is particularly useful for moving Skuid data from a sandbox—or testing space—to production for end-users. Whether that be Skuid on Salesforce—which takes [some additional deployment steps](https://docs.skuid.com/latest/en/skuid/deploy/salesforce/org-to-org.html)—or the nearly complete transferral possible for Skuid Platform sites.

The `-u`, `-p` and `--host` flags can be used to temporarily set different authentication and platform settings.

For example, consider a workflow like below:

```bash
# Set the authentication credentials for the sandbox site, assuming they are not set already
export SKUID_UN='{sandbox-username}'
export SKUID_PW='{sandbox-password}'
export SKUID_HOST='{https://example-sandbox.skuidsite.com}'
# Retrieve sandbox data
skuid retrieve
# Deploy sandbox data to production by temporarily using different credentials
skuid deploy -u production-username -p production-password  --host https://production-sandbox.skuidsite.com}
```

Consider automating this process using shell scripts. Experiment to find the workflow that best suits your lifecycle management processes.

## Troubleshooting

`skuid` tries to only show basic information to avoid cluttering the terminal. If you're seeing errors, a good first step is try the command again with the `-v` or `--verbose` flag to log more information.

Some possible error messages include:

- `skuid push` returns `Pushing 0 pages` and does not push my Skuid pages.
  - This means there are no Skuid pages within the current directory. You must navigate to a directory that contains pages or use the `--dir` flag, e.g. `skuid push --dir skuidpages`.
- `unexpected end of JSON input`
  - This error can appear when attempting to deploy to a Skuid Platform site. This may mean that the `SKUID_UN` and `SKUID_PW` variables are not set.
- `Error deploying metadata: Post https://example.skuidsite.com/api/v1/metadata/deploy: EOF`
  - This may mean your Skuid credentials are not correct within your environment variables. Verify both your username and password.
- `Post example.my.salesforce.com/services/oauth2/token: unsupported protocol scheme ""`
  - The `SKUID_HOST` variable or `--host` flag value may not have the `https` protocol in the proper place.

    **Correct**

    ```bash
    export SKUID_HOST=https://example.my.salesforce.com
    ```

    **Incorrect**

    ```bash
    export SKUID_HOST=example.my.salesforce.com
    ```

- `Post https://example.my.salesforce.com/oauth2/token: dial tcp: lookup https://example.my.salesforce.com:no such host` or `invalid character '<' looking for beginning of value`
  - This indicates the Skuid site or Salesforce org that the `SKUID_HOST` value points to may not exist. Ensure that you've correctly enter the URL for your platform.
- `Get /services/apexrest/skuid/api/v1/pages?module=: unsupported protocol scheme ""`
  - You may not have one of your authentication variables set appropriately. Verify the username, password, host, client ID, and client secret variables are set appropriately for your Salesforce org.
    - If in a **macOS** or **Linux** environment, ensure the user credentials are encased within **single quotes**: 

      **Correct**

      ```bash
      export SKUID_UN='username'
      ```

      **Incorrect**
      
      ```bash
      export SKUID_UN=username
      ```

      Some shells may incorrectly interpret characters within credentials without these quotes.

  - Ensure that both your Salesforce password _and_ your security token are set—written together with no characters between them— within the `SKUID_PW` variable.
  - The user credentials may not have permission to access the connected app within Salesforce. Ensure the credentials have access either through [the user profile](https://help.salesforce.com/articleView?id=admin_userprofiles.htm&type=5) or [a Salesforce permission set](https://help.salesforce.com/articleView?id=perm_sets_overview.htm&type=5), and also ensure users are [API enabled](https://help.salesforce.com/articleView?id=admin_userperms.htm&type=5). 
- `Error retrieving metadata:  Error making HTTP request%!(EXTRA string=413 Payload Too Large)`
  - This can indicate that there are files besides Skuid data within the directory you are attempting to deploy. Verify that there are no other applications or files in the Skuid data directory. If you are using the `skuid` CLI application in the same directory as your data, you'll need to move your data a different directory and use the `--dir` flag to point to it.
- `Error retrieving metadata:  Error making HTTP request%!(EXTRA string=401 Unauthorized)`
  - Your `SKUID_UN` or `SKUID_PW` may not be set correctly. Verify your user credentials.
  - **Ensure you have the latest version of the `skuid` CLI.** If your credentials are correct, try removing your current installation and [reinstalling the binary.](#installation)
- `Error deploying metadata: Error making HTTP request%!(EXTRA string=400 Bad Request)` or `Error deploying metadata: Error making HTTP request%!(EXTRA string=500 Internal Server Error)`
  - These errors indicate there may be an issue with the retrieval of [your Skuid Platform apps.](https://docs.skuid.com/latest/en/skuid-platform/deploying-apps-in-skuid-platform.html#apps)
    - Ensure that all apps you have retrieved are properly configured, with correct routes and corresponding pages for those routes.
    - Verify there are no issues with [user profile access](https://docs.skuid.com/latest/en/skuid-platform/user-and-permission-management.html#profiles) to the apps you are attempting to deploy.
    - If all the options appear correct, **delete the `apps` folder** on your local machine and attempt `skuid deploy` again. Recreate your apps manually on the destination Platform site.
- `Error executing retrieve plan: Post https://<IP address>/api/v2/metadata/retrieve: dial tcp <IP address>: i/o timeout`
  - This error indicates an issue retrieving metadata from a **data service**. Verify the URLs/IP addresses of all data services listed in **Configure > Data Sources > Data Services** are correct, and ensure that they all can respond to Skuid's metadata requests.
- `"{\"error\":\"You must provide a \\\"module\\\" URL Parameter containing the names of the Modules of Pages (comma-separated) AND/OR a \\\"page\\\" URL Parameter containing the names of the Pages (comma-separated) you would like to retrieve.\"}" json: cannot unmarshal string into Go value of type types.PullResponse`
  - Some versions of Skuid and the `skuid` CLI do not allow for pulling pages without specifying a module. For maximum compatibility, ensure you are on Skuid v11.2.12 and above, as well as `skuid` CLI 0.3.3 or above.

## skuid vs skuid-grunt

[`skuid-grunt`](https://github.com/skuid/skuid-grunt) previously provided the above-documented push/pull functionality as a Grunt plugin. While great for projects already using NodeJS and Grunt, the plugin was problematic if you didn't want to require those dependencies. `skuid` solves that dependency problem by producing a self-contained CLI to perform all the same operations.
