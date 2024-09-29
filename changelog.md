# Changelog

## [0.0.2] - (Unreleased)

### Added
- Use the default gpg key from HashiCorp for official providers


## [0.0.1] - 2024-09-29

### Added
- Initial release of `tf-provider-upload` tool.
- Support for uploading provider files to Terraform Enterprise/Cloud private registry.
- Added option to specify `gpg_key` for signing provider files.
- Added support for specifying the `organization` and `providerfolder` options.
- Default configurations for `gpg_key`, `organization`, and `providerfolder` are provided.
- Option to define `tfeHostname` for the target Terraform Enterprise server.
- Support for `token` configuration via environment variable (`TFE_TOKEN`) or CLI option.
- Added `version` flag to display the tool's version.

### Features
- Upload provider files securely using a GPG key.
- Batch uploads of provider files from a specified folder.
- Easy integration with Terraform Enterprise or Cloud.
