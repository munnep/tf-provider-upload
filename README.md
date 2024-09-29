# tf-provider-upload

tf-provider-upload is a command-line tool that simplifies uploading Terraform providers to the private registry of Terraform Enterprise (TFE) or Terraform Cloud.

## Features:

- Provider File Upload: Upload your Terraform providers to the private registry in TFE/TFC.
- GPG Signing: Sign provider files with a GPG key before uploading.

## Installation:
**Pre-built binaries:**

Download the pre-built binary for your operating system from [link to releases page](https://github.com/munnep/tf-provider-upload/releases).

**From source:**

Clone this repository and build the binary:

```
git clone https://github.com/your-username/tf-provider-upload.git
cd tf-provider-upload
go build -o tf-provider-upload
```

## Usage:

The tool can be invoked from the command line with various options:
```
./tf-provider-upload -h
```

Options:

- gpg_key string  
  GPG key string to sign the provider files with (default "<this must be set>"). See [here](gpg-key/README.md) how to create a key if you do not have on yet

- organization string  
  Organization in the Terraform Enterprise (default "<this must be set>")

- providerfolder string  
  Relative directory path where the provider files are stored (default "<this must be set>")

- tfeHostname string  
  Hostname of the Terraform Enterprise (default "<this must be set>")

- token string  
  By default, it will look for the environment variable TFE_TOKEN

- version  
  Show the version of the tool and exit.

Example Command:

```
./tf-provider-upload \
  -gpg_key "222AFF6CAD7A1CF67A197DA296031B182E25BF7A" \
  -organization "my-organization" \
  -providerfolder "path/to/providers" \
  -tfeHostname "tfe.mycompany.com"
```

### Setting the TFE_TOKEN:

Ensure the TFE_TOKEN environment variable is set for authentication if not passed directly via the -token option:
```
export TFE_TOKEN="your-tfe-token"
```

# How to Use:

1. Prepare your custom provider files in the folder specified by the -providerfolder option.
The file format should look like the following
```
terraform-provider-azurerm_3.107.0_darwin_amd64.zip
terraform-provider-azurerm_3.107.0_linux_amd64.zip
```
2. Set your TFE authentication token via the -token option or the TFE_TOKEN environment variable.
3. Run the tool to upload the provider to your Terraform Enterprise or Terraform Cloud private registry.

# License:

This project is licensed under the MIT License. See the LICENSE file for more details.
