# Generating a GPG Key Using Batch Mode for tf-provider-upload

This guide explains how to generate a GPG key using the `--batch` function in GnuPG. The GPG key will be used to sign provider files for the `tf-provider-upload` tool.

## GPG Batch File Configuration

Create a file named `gpg-key.batch` with the following content. This file contains the necessary parameters to generate a 4096-bit RSA key pair without passphrase protection:

```
Key-Type: 1
Key-Length: 4096
Subkey-Type: 1
Subkey-Length: 4096
Name-Real: tf-provider-upload
Name-Email: tf-provider-upload@example.com
Expire-Date: 0
%no-protection
%commit
```

Explanation of fields:

- `Key-Type: 1`: Specifies the use of RSA encryption.
- `Key-Length: 4096`: Sets the key length to 4096 bits.
- `Subkey-Type: 1`: Specifies the subkey also uses RSA encryption.
- `Subkey-Length: 4096`: Subkey length is set to 4096 bits.
- `Name-Real`: The name associated with the key (`tf-provider-upload`).
- `Name-Email`: The email associated with the key (`tf-provider-upload@example.com`).
- `Expire-Date: 0`: The key will never expire.
- `%no-protection`: No passphrase protection is used on the private key.
- `%commit`: Finalizes the key generation process.

## Generate the GPG Key

To generate the GPG key, use the following command in the terminal, specifying the `gpg-key.batch` file created above:

```
gpg --batch --generate-key gpg-key.batch
```
This command will generate a 4096-bit RSA key pair as defined in the batch file.

## List the Generated GPG Key

Once the key is generated, you can list it using the following command:

```
gpg --list-keys
```
You will see an output similar to this:

```
/home/user/.gnupg/pubring.kbx
-----------------------------
pub   rsa4096 2024-09-27 [SC] [expires: never]
      222AFF6CAD7A1CF67A197DA296031B182E25BF7A
uid           [ultimate] tf-provider-upload <tf-provider-upload@example.com>
sub   rsa4096 2024-09-27 [E] [expires: never]
```

The key ID is the 16-character string, which in this example is `222AFF6CAD7A1CF67A197DA296031B182E25BF7A`. Use this key ID when signing provider files with `tf-provider-upload`.

## Export the GPG Public Key

If you need to export the public key to share it or use it on another system, run:

```
gpg --armor --export tf-provider-upload@example.com > tf-provider-upload-public.key
```

This will create an ASCII-armored file containing the public key.

## Importing the Private Key

To back up and import the private key on another system, export it first:

```
gpg --armor --export-secret-keys tf-provider-upload@example.com > tf-provider-upload-private.key
```

On the new system, you can import it like this:

```
gpg --import tf-provider-upload-private.key
```

## Usage in tf-provider-upload

You can now use the generated GPG key to sign provider files in `tf-provider-upload`. Simply provide the key ID when prompted, or set it in the configuration.
```
./tf-provider-upload -gpg_key "222AFF6CAD7A1CF67A197DA296031B182E25BF7A"
```

## License

This project is licensed under the MIT License. See the LICENSE file for more details.
