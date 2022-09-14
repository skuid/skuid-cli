# **Skuid CLI CDK**: Tides 

![GitHub go.mod Go version (subdirectory of monorepo)](https://img.shields.io/github/go-mod/go-version/skuid/skuid-cli) ![GitHub release (latest by date)](https://img.shields.io/github/v/release/skuid/skuid-cli)
![Supported Operating Systems](https://img.shields.io/badge/os-mac-brightgreen)
![Supported Operating Systems](https://img.shields.io/badge/os-linux-brightgreen)
## Description

**Skuid** `tides` is a command-line, cloud development kit. This package is meant to help with 
- the retrieving of **Skuid** objects and apps (data and metadata) 
- the deployment of **Skuid** objects and apps

across **Skuid** Platform sites. 

## Aim

The aim of **Skuid** `tides` is to help with the management of **Skuid** sites. 

A common use case is to have **Skuid** `tides` assist as a migration tool from _developmet_ to _production_ **Skuid** sites. 

## Looking for Skuid CLI for Salesforce?

We now have [`tides-sfdx`](https://github.com/skuid/skuid-sfdx), our open-source Salesforce CLI plugin designed to handle Skuid metadata on the Salesforce platform.
## Supported Operating Systems
![Supported Operating Systems](https://img.shields.io/badge/os-mac-brightgreen)
![Supported Operating Systems](https://img.shields.io/badge/os-linux-brightgreen)


## Installation

### macOS and Linux

If you have `curl`, `grep` and `awk` in your environment, you can quickly install the application via the command line:

<!-- these commands don't work:
```bash
# macOS:
curl -Lo tides $(curl -sL https://api.github.com/repos/skuid/skuid-cli/releases/latest | grep "browser_download_url.*darwin" | awk -F '"' '{print $4}')

# Linux:
# curl -Lo skuid $(curl -sL https://api.github.com/repos/skuid/skuid-cli/releases/latest | grep "browser_download_url.*linux" | awk -F '"' '{print $4}')

# Give the skuid application the permissions it needs to run
chmod +x tides

# Move the skuid application to a folder where it can be used easily, 
# for example, a directory in your $PATH.
# Enter your computer account password if prompted
sudo mv tides /usr/local/bin/tides
```
-->

To manually install the application, follow these steps:

1. Download [the latest release of the `tides` application binary.](https://github.com/skuid/skuid-cli/releases)
1. Navigate to the directory containing the `tides` binary file in a terminal application.

   ```bash
   cd /path/to/the/downloaded/binary
   ```

1. Rename the downloaded application binary file to `tides`:

   ```bash
   # macOS:
   mv tides_darwin_amd64 tides
   # Linux:
   # mv tides_linux_amd64 tides
   ```

1. Give the application executable permissions:

   ```bash
   chmod +x tides
   ```

1. Move the application to a folder in your `$PATH` variable, like `/usr/local/bin`, or add the application's folder to the PATH variable:

   ```bash
   mv tides /usr/local/bin/tides
   ```

1. Verify that you can run the application:

   ```bash
   tides --help
   ```
<!-- 

let's... not support windows.

### Windows

1. Download the [latest Windows executable release](https://github.com/skuid/skuid-cli/releases) in your web browser.
1. (_Optional_) Move the executable to a more permanent location, such as `C:\Program Files\Skuid\`.
1. (_Optional_) Set an alias to easily access the executable.

   In Windows Powershell, use Set-Alias. For information about saving aliases in Powershell, see [Microsoft documentation](https://docs.microsoft.com/en-us/powershell/module/microsoft.powershell.utility/set-alias?view=powershell-6)

   ```bash
   Set-Alias tides C:\Path\To\tides_windows_amd64.exe
   ```

   In the Windows Command Prompt, use doskey. For information about saving doskey aliases for future sessions, see [Microsoft documentation](https://technet.microsoft.com/en-us/library/ff382652.aspx).

   ```bash
   doskey tides=C:\Path\To\tides_windows_amd64.exe
   ```
	 
	 -->

1. Invoke the executable by typing its name or alias and pressing enter: `tides --help`.

#### Building from source

To build the application from the source, first clone the repository or download the source. You also need [Go installed on your machine](https://golang.org/doc/install).

- To build from source for your machine:

  ```bash
  go build .
  ```
	
- You should now see an executable for your machine.

	```bash
	./tides
	```

### Upgrading skuid CLI

To upgrade `tides`, first remove the previous version's binary:

```bash
# Use the rm command with the appropriate path to the skuid binary
# For example, if the binary is installed in /usr/local/bin:
rm /usr/local/bin/skuid
```

After removing the previous version, repeat [the installation steps listed above.](#installation)

## Configuration

By default, `tides` uses environment variables to provide credentials for interacting with Skuid APIs under the hood. While you can set username, password, host, and connected app values with flags, consider setting the following environment variables to avoid entering credentials with every command.

Which environment variables you'll need to set depends on which platform you'll be connecting to.

### Skuid Platform

#### macOS and Linux

Enter the appropriate information in the format listed below, listing your own username, password, etc., immediately following the `=` equals sign. You can drop these in your `~/.bash_profile`, `~/.zshrc`, or into a `.env` file in the project directory.

### `.zprofile` / `.bash_profile` / `.env`
```bash
# essentials
export SKUID_PW
export SKUID_UN
export SKUID_HOST

# advanced
export SKUID_LOGGING
export SKUID_VERBOSE
export SKUID_LOGGING_FOLDER
export SKUID_FILE_LOGGING
export SKUID_LOGGING_DIRECTORY
export SKUID_DEFAULT_FOLDER
```

<!-- screw windows!
#### Windows

How you set your environment variables differs depending on your command line program of choice:

**Powershell**

```bash
Set-Item Env:tides_UN 'username'
Set-Item Env:tides_PW 'password'
Set-Item Env:tides_HOST 'https://example.skuidsite.com'
```
**Command Prompt**

```bash
Set tides_UN=username
Set tides_PW=password
Set tides_HOST=https://example.skuidsite.com
```
-->

## Usage

`tides` is used to do two things:

1. **RETRIEVE** Skuid data and metadata _from_ Skuid NLX and **store** it on the local machine
1. **DEPLOY** Skuid data and metadata _from_ the local machine _to_ Skuid NLX

All of this is done using the following syntax:

```bash
  tides [command] [flags]
```

**For the latest documentation, PLEASE run the command `tides --help` or `tides [command] --help` for more information.

## Skuid Object Structure

When pulling from the Salesforce platform, all pages are stored within a `pages` directory in the current working directory. Each page consists of two files:

- A `.json` file, consisting of the page's metadata
- The page's `.xml` file

When retrieving from Skuid NLXs Platform sites, the following is downloaded:

- All [apps, and the routes within them](https://docs.skuid.com/latest/en/skuid-platform/deploying-apps-in-skuid-platform.html), in the `apps` directory
- All [authentication providers](https://docs.skuid.com/latest/en/skuid/metadata-objects/v1/authprovider.html) in the `authproviders` directory
- All [component packs](https://docs.skuid.com/latest/en/skuid/components/original/build-component-packs.html) in the `componentpacks` directory
- All [data services](https://docs.skuid.com/latest/en/data/private-data-service/#create-the-data-service) in the `dataservices` directory (Note: Data Services are only available if you have the "Private Data Service" feature enabled for your site. Contact your Skuid representative for more information.)
- All [data sources](https://docs.skuid.com/latest/en/skuid/metadata-objects/v1/datasource.html) in the `datasources` directory
- All [design systems](https://docs.skuid.com/latest/en/skuid/metadata-objects/v1/designsystem.html) in the `designsystems` directory
- All [files](https://docs.skuid.com/latest/en/skuid/metadata-objects/v1/file.html) in the `files` directory
- All [pages](https://docs.skuid.com/latest/en/skuid/metadata-objects/v1/page.html) in the `pages` directory
- All [profiles](https://docs.skuid.com/latest/en/skuid-platform/user-and-permission-management.html#profiles) in the `profiles` directory
- All [site settings](https://docs.skuid.com/latest/en/skuid/metadata-objects/v1/site.html) in the `site` directory (including a site's logo and/or favicon)
- All [themes](https://docs.skuid.com/latest/en/skuid/metadata-objects/v1/theme.html) in the `themes` directory
- All [variables] in the `variables` directory

### What Is Not Retrieved by skuid

- **Created by** and **Modified by** metadata for pages
  - When deploying Skuid pages to a site, the created/modified by user and date matches **the identity of whoever is running the deployment**, as well as **the time and date of the deployment.**
- **Skuid Platform:**
  - Authentication provider credentials
    - **You must re-enter any client ID and client secret pairs on all Skuid authentication providers in the target site**, even if those authentication providers already existed.
  - Users and user data for Skuid Platform
    - While user profiles are transferred, individual user accounts and their information are not. Users must be manually re-created—or [provisioned through single sign-on](https://docs.skuid.com/latest/en/skuid/single-sign-on/#user-provisioning-within-skuid-platform)—for at least the first deployment.
## Troubleshooting

`tides` tries to only show basic information to avoid cluttering the terminal. 

If you're seeing errors, a good first step is try the command again with the `-v` or `--verbose` flag to log more information.

### Logging Levels

There are three logging levels with corresponding flags

- `--verbose`
- `--trace`
- `--diagnostic`

Which increase in detail (and increased difficulty for legibility). It is recommended to use `--verbose` unless you are an advanced user.