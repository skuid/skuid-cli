# skuid

`skuid` is a command line application (CLI) for retrieving and deploying Skuid objects (data and metadata) on both Skuid Platform sites and Salesforce orgs running the Skuid managed package.

While Skuid is a cloud user experience platform, it can be helpful to:

- download an entire site's worth of pages to make small adjustments locally
- move Skuid configurations from a sandbox site to a production site
- store Skuid configurations in a version control system (VCS)

While it's possible to save page XML and create page packs, moving entire apps' worth of Skuid objects from site to site can prove challenging.

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
# Download the skuid application
# On a macOS machine? Use this:
wget https://github.com/skuid/skuid-cli/releases/download/3/skuid_darwin_amd64 -O skuid
# On a Linux machine? Use this instead:
# wget https://github.com/skuid/skuid-cli/releases/download/0.3.10/skuid_linux_amd64 -O skuid
# Give the skuid application the permissions it needs to run
chmod +x skuid
# Move the skuid application to a folder where it can be used easily
# Enter your computer account password if prompted
sudo mv skuid /usr/local/bin/skuid
```

To manually install the application, follow these steps:

1. Download [the latest release of the `skuid` application binary.](https://github.com/skuid/skuid-cli/releases)
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

   _Note_: If you choose to update your shell profile you'll need to `source` your shell profile or restart your session for those changes to take effect.

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

## Making a new release

In order to make a release for each OS, you'll first want to checkout master and update the version number in `Makefile`.

Then you can simply checkout master and run the `release` make target.

```back
make release
```

This will tag the current master commit with the version number you specified in `Makefile` and build the binaries for use on each platform.

You'll push that new tag to github:

```bash
git push <remote name> <tag_name>
```

Once the tag is pushed, you should be able to add a new release referencing that tag at https://github.com/skuid/skuid-cli/releases. When adding the new release, there is a section for uploading any binaries for the release which you will use to add the newly built binaries you just created.
