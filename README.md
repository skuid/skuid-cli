# skuid

`skuid` is a command line application (CLI) for retrieving and deploying Skuid objects (data and metadata) on both Skuid Platform sites and Salesforce orgs running the Skuid managed package.

While Skuid is a cloud user experience platform, it can be helpful to:

- download an entire site's worth of pages to make small adjustments locally
- move Skuid configurations from a sandbox site to a production site
- store Skuid configurations in a version control system (VCS)

While it's possible to save page XML and create page packs, moving entire apps of Skuid objects from site to site can prove challenging.

Enter the `skuid` CLI. Using `skuid`, you can easily pull—download—Skuid pages and push—upload—Skuid metadata from one site to another using only a few commands.

* [Prerequisites](#prerequisites)
* [Installation](#installation)
	* [Building from source](#building-from-source)
* [Usage](#usage)
* [skuid vs skuid-grunt](#skuid-vs-skuid-grunt)

## Prerequisites

- Some basic knowledge of the command line is required. You should be able to navigate the file system with `cd` and enter the `skuid` command.
- If pulling and pushing pages with Skuid on Salesforce...
  - [Enable My Domain](https://help.salesforce.com/articleView?id=domain_name_overview.htm&type=5) for your Salesforce org
  - Configure a [Salesforce connected app](https://help.salesforce.com/articleView?id=connected_app_overview.htm&type=5) and have both the _consumer key_ and the _consumer secret_ available

## Installation

### macOS and Linux

To quickly install the application, copy and paste the following commands in the terminal:

```bash
# On a macOS machine? Use this:
wget https://github.com/skuid/skuid/releases/download/0.2.0/skuid_darwin_amd64 -O skuid
# On a Linux machine? Use this instead:
# wget https://github.com/skuid/skuid/releases/download/0.2.0/skuid_linux_amd64 -O skuid
chmod +x skuid
sudo mv skuid /usr/local/bin/skuid
```

To manually install the application, follow these steps:

1. Download [the latest release of the `skuid` application binary.](https://github.com/skuid/skuid/releases)
1. Navigate to the directory containing the `skuid` binary file in a terminal application.

   ```bash
   cd a-folder
   ```

1. Rename the downloaded application binary file to `skuid`:

   ```bash
   mv skuid_darwin_amd64 skuid
   # or for Linux:
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

1. Verify that you can run the application:

   ```bash
   skuid --help
   ```

### Windows

1. Download the [latest Windows executable release](https://github.com/skuid/skuid/releases) in your web browser.
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

Use the [the Go programming language?](https://golang.org/doc/install) If so, you can also install `skuid` by running `go get github.com/skuid/skuid`.

### Building from source

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

## Usage

For up-to-date usage information see [our documentation page.](https://docs.skuid.com/latest/en/skuid/cli)

## skuid vs skuid-grunt

[`skuid-grunt`](https://github.com/skuid/skuid-grunt) previously provided this push/pull functionality as a Grunt plugin. While great for projects already using NodeJS and Grunt, the plugin was not so great if you didn't want to require those dependencies. `skuid` solves that dependency problem by producing a self-contained CLI to perform all the same operations.