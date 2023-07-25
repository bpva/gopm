# gopm — Go Package Manager

The Go package manager is designed to perform the following tasks:

- Package files into an archive and upload them to a server via SSH.
- Download archive files via SSH and unpack them.
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

## Commands
The package manager will provide the following commands:

- `gopm create ./packet.json`: Packages the files specified in the package file into an archive.
- `gopm update ./packages.json`: Downloads archive files via SSH and unpacks them.

And I could make any reasonable assumptions to simplify the development.