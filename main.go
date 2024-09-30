package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"tf-upload-provider/gpg"
	"tf-upload-provider/providerupload"
)

const version = "0.0.1"

func main() {

	providerfolderPtr := flag.String("providerfolder", "", "relative directory path where the provider files are stored")
	gpg_keyPtr := flag.String("gpg_key", "", "gpg key string to sign the provider files with")
	tokenPtr := flag.String("token", "", "By default it will look for the environment variable TFE_TOKEN")
	organizationyPtr := flag.String("organization", "", "Organization in the terraform enterprise")
	tfeHostnamePtr := flag.String("tfeHostname", "", "Hostname of the terraform enterprise")

	
	versionPtr := flag.Bool("version", false, "Show the version")
	
	flag.Parse()
	
	// check if the token is set as environment variable
	if *tokenPtr == "" {
		*tokenPtr = os.Getenv("TFE_TOKEN")
		if *tokenPtr == "" {
			fmt.Println("Error: TFE_TOKEN is not set or is empty. Exiting...")
			os.Exit(1)
		}
	}

	// print the version and exit
	if *versionPtr {
		fmt.Println(version)
		os.Exit(0)
	}

	fmt.Println("Upload provider to private registry")

	// Create a directory where all files during the running of the tool are stored
	tmp_directory()

	// generate the public gpg key as string
	gpgPublicKeyString := gpg.Keys(*providerfolderPtr, *gpg_keyPtr)

	// check if the gpg key is already uploaded in terraform enterprise
	keyExists, gpgKeyID := gpg.CheckGPGKey(gpgPublicKeyString, *tokenPtr, *tfeHostnamePtr, *organizationyPtr)
	// fmt.Println(keyExists, gpgKeyID)

	// if not already uploaded do this
	if !keyExists {
		fmt.Println("PGP public signature uploading")
		gpgKeyID = gpg.UploadGPGKey(gpgPublicKeyString, *tokenPtr, *tfeHostnamePtr, *organizationyPtr)
	} else {
		fmt.Println("PGP public signature already uploaded")
	}
	// fmt.Println(gpgKeyID)
	// fmt.Println(gpgKeyID)

	// Get the providerName and providerVersion
	providerName, providerVersion, err := gpg.GetProviderNameVersion(*providerfolderPtr)
	if err != nil {
		log.Fatal(err)
	}

	//providerupload
	providerupload.CreateProvider(gpgKeyID, *tokenPtr, *tfeHostnamePtr, *organizationyPtr, providerName)

	//providerVersionUpload
	shasumsUpload, shasumsSigUpload := providerupload.CreateVersionProvider(gpgKeyID, *tokenPtr, *tfeHostnamePtr, *organizationyPtr, providerName, providerVersion)

	// upload shasum files
	providerupload.ShaSumUpload(shasumsUpload, shasumsSigUpload, *providerfolderPtr)

	providerupload.UploadProviderVersionPlatform(*providerfolderPtr, *tokenPtr, *tfeHostnamePtr, *organizationyPtr)

}

func tmp_directory() {
	err := os.Mkdir(".tf-provider-upload", 0750)
	if err != nil && !os.IsExist(err) {
		log.Fatal(err)
	}
}
