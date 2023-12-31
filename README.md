# gopm — Package Manager written in Go

The **gopm** is designed to perform the following tasks:

- Package files into an archive and upload them to a server via SSH.
- Download archive files via SSH and unpack them.
## Installation

You can install the Go Package Manager using the following methods:

### Method 1: Using `go install`

You can install the latest version of the Go Package Manager by running the following command in your terminal:

`go install github.com/bpva/gopm/cmd/gopm@latest`

Then simply run:

`gopm`


### Method 2: From the Releases Page

Alternatively, you can download the desired release version of the Go Package Manager from the Releases page (https://github.com/bpva/gopm/releases) on GitHub. Choose the appropriate binary for your operating system and architecture, and then follow the installation instructions provided in the release documentation.
## Configuration

To configure the tool, you can use a `.env` file or environment variables. The tool supports the following configuration options:

- `GOPM_SSH_MODE`: The SSH mode to use. Set it to `login+password` for login and password authentication, or `key` for key-based authentication.
- `GOPM_SSH_LOGIN`: The SSH login username.
- `SSH_KEY_PATH`: The path to the private key file for key-based authentication. Leave it empty if using login and password authentication.
- `GOPM_SSH_PASSWORD`: The SSH login password. Leave it empty if using key-based authentication.
- `GOPM_SSH_HOST`: The SSH host to connect to.
- `GOPM_SSH_PORT`: The SSH port to use (default: `22`).

### Using the `.env` file

To use the `.env` file, create a file named `.env` in the root directory of your project. The file should follow the key-value pair format, where each line represents a configuration option in the format `KEY=VALUE`. Example can be found in root directory as example.env (rename it to .env)

### Using Environment Variables

Alternatively, you can set the configuration options directly using environment variables. Ensure that the required environment variables are set with the appropriate values.

### Specifying the `.env` File Location

If you want to specify a different location for the `.env` file, you can use the `-env` flag when running the tool. For example:
```shell
gopm create testdata/package.json -env /path/to/.env
```
## Usage
The package manager will provide the following commands:

- `gopm create ./packet.json`: Packages the files specified in the package file into an archive.
- `gopm update ./packages.json`: Downloads archive files via SSH and unpacks them.

## Package File Format
The package file should have either a `.yaml` or `.json` format. It should include paths to select files using glob patterns.

## Example Package File:
**packet.json**

```json
{
  "name": "packet-1",
  "ver": "1.10",
  "targets": [
    "./archivethis1/*.txt",
    {"path": "./archivethis2/", "exclude": "*.tmp"}
  ],
  "packets": [
    {"name": "packet-3", "ver": "<=2.0"}
  ]
}
```

## Example Package File for Unpacking:
**packages.json**

```json
{
  "packages": [
    {"name": "packet-1", "ver": ">=1.10"},
    {"name": "packet-2"},
    {"name": "packet-3", "ver": "<=1.10"}
  ]
}
 ```

And I could make any reasonable assumptions to simplify the development.
