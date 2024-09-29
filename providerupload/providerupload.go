package providerupload

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	// "mime/multipart"
	"net/http"
	"os"
	"strings"
)

type ProviderPayload struct {
	Data struct {
		Type       string `json:"type"`
		Attributes struct {
			Name         string `json:"name"`
			Namespace    string `json:"namespace"`
			RegistryName string `json:"registry-name"`
		} `json:"attributes"`
	} `json:"data"`
}

type ProviderVersionPayload struct {
	Data struct {
		Type       string `json:"type"`
		Attributes struct {
			Version   string `json:"version"`
			KeyID     string `json:"key-id"`
			Protocols string `json:"protocols"`
		} `json:"attributes"`
	} `json:"data"`
}

func CreateProvider(gpgKeyID string, token string, tfeHostname string, organization string, providerName string) {
	// fmt.Println("providerupload", gpgKeyID)

	// fmt.Println(gpgPublicKeyString)
	// Create the payload struct
	payload := ProviderPayload{}
	payload.Data.Type = "registry-providers"
	payload.Data.Attributes.Name = providerName
	payload.Data.Attributes.Namespace = organization
	payload.Data.Attributes.RegistryName = "private"

	// Marshal the struct to JSON
	payloadBytes, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		log.Fatal("Error marshaling JSON payload:", err)
	}

	// fmt.Println(string(payloadBytes))

	// fmt.Println(string(payloadBytes))
	// Create a request body from the payload
	requestBody := bytes.NewReader(payloadBytes)

	// Construct the URL
	myURL := "https://" + tfeHostname + "/api/v2/organizations/" + organization + "/registry-providers"

	// Create a new POST request
	req, err := http.NewRequest("POST", myURL, requestBody)
	if err != nil {
		log.Fatal("Error creating request:", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/vnd.api+json")

	// Send the request
	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		log.Fatal("Error sending request:", err)
	} else {
		fmt.Println("Provider private registry created")
	}
	defer response.Body.Close()

	// TODO
	// WHAT IF THE VERSION ALREADY EXIST? AT THE MOMENT WE IGNORE THE MESSAGE ERROR

	// Read and print the response body
	// content, err := io.ReadAll(response.Body)
	// if err != nil {
	// 	log.Fatal("Error reading response body:", err)
	// }

	// fmt.Println("print content", string(content))
	// Unmarshal the JSON response into the struct
	// var responsePayload PayloadResponse
	// err = json.Unmarshal(content, &responsePayload)
	// if err != nil {
	// 	log.Fatal("Error unmarshaling JSON response:", err)
	// }

	// // Extract the key-id
	// gpgKeyID := responsePayload.Data.Attributes.KeyID

	// Return the key-id
	return
}

func CreateVersionProvider(gpgKeyID string, token string, tfeHostname string, organization string, providerName string, providerVersion string) (shasumsUpload string, shasumsSigUpload string) {
	// fmt.Println("providerupload", gpgKeyID)

	// fmt.Println(gpgPublicKeyString)
	// Create the payload struct
	payload := ProviderVersionPayload{}
	payload.Data.Type = "registry-provider-versions"
	payload.Data.Attributes.Version = providerVersion
	payload.Data.Attributes.KeyID = gpgKeyID
	payload.Data.Attributes.Protocols = "5.0"

	// Marshal the struct to JSON
	payloadBytes, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		log.Fatal("Error marshaling JSON payload:", err)
	}

	// fmt.Println(string(payloadBytes))
	// Create a request body from the payload
	requestBody := bytes.NewReader(payloadBytes)

	// Construct the URL
	myURL := "https://" + tfeHostname + "/api/v2/organizations/" + organization + "/registry-providers/private/" + organization + "/" + providerName + "/versions"
	// fmt.Println(myURL)

	// Create a new POST request
	req, err := http.NewRequest("POST", myURL, requestBody)
	if err != nil {
		log.Fatal("Error creating request:", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/vnd.api+json")

	// Send the request
	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		log.Fatal("Error sending request:", err)
	} else {
		fmt.Println("Provider version created")
	}
	defer response.Body.Close()

	// TODO
	// WHAT IF THE VERSION ALREADY EXIST? AT THE MOMENT WE IGNORE THE MESSAGE ERROR

	// Read and print the response body
	content, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatal("Error reading response body:", err)
	}

	// fmt.Println("print content", string(content))

	var versionResult map[string]interface{}

	// Unmarshal JSON into a map which is a different approach then knowing the structure. Just as a practice
	err = json.Unmarshal(content, &versionResult)
	if err != nil {
		fmt.Println("error: ", err)
		return
	}
	// fmt.Println("Received JSON:", versionResult)

	// Navigate through the nested structure to find the `shasums-upload` and `shasums-sig-upload`
	if data, ok := versionResult["data"].(map[string]interface{}); ok {
		if links, ok := data["links"].(map[string]interface{}); ok {
			// Extract shasums-upload
			if su, ok := links["shasums-upload"].(string); ok {
				shasumsUpload = su
			}

			// Extract shasums-sig-upload
			if ssu, ok := links["shasums-sig-upload"].(string); ok {
				shasumsSigUpload = ssu
			}
		}
	}

	if shasumsUpload == "" {
		fmt.Println("error: Version already created")
	}

	// if shasumsSigUpload == "" {
	// 	fmt.Println("error: empty sig upload string")
	// }

	// fmt.Println("shasum-upoad:", shasumsUpload )
	// fmt.Println("shasum-sig-upload:",shasumsSigUpload )

	return shasumsUpload, shasumsSigUpload
}

func ShaSumUpload(shasumsUpload string, shasumsSigUpload string, providerfolder string) {

	if shasumsUpload == "" {
		fmt.Println("error: No location to upload")
		return
	}

	// Upload SHA256SUMS file
	shaSumfileName := providerfolder + "/files.SHA256SUMS"
	if err := uploadFile(shasumsUpload, shaSumfileName); err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("SHA256SUMS file uploaded successfully")

	// Upload SHA256SUMS.sig file
	shaSumSigfileName := providerfolder + "/files.SHA256SUMS.sig"
	if err := uploadFile(shasumsSigUpload, shaSumSigfileName); err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("SHA256SUMS.sig file uploaded successfully")
}

type ProviderVersionPayloadPlatform struct {
	Data struct {
		Type       string `json:"type"`
		Attributes struct {
			Os       string `json:"os"`
			Arch     string `json:"arch"`
			Shasum   string `json:"shasum"`
			Filename string `json:"filename"`
		} `json:"attributes"`
	} `json:"data"`
}

// UploadProviderVersionPlatform is the main function to process the files
func UploadProviderVersionPlatform(providerfolder, token, tfeHostname, organization string) {
	// Construct the path to the SHA256SUMS file
	filename := providerfolder + "/files.SHA256SUMS"

	// Process the file
	if err := processFile(filename, token, tfeHostname, organization, providerfolder); err != nil {
		fmt.Println("Error processing file:", err)
		return
	}

	// Here you would add code to upload the files, if necessary
	// For example:
	// uploadProviderVersion(shasumsUpload, shasumsSigUpload, providerfolder)
}

// processFile reads the file and parses each line
func processFile(filepath, token, tfeHostname, organization, providerfolder string) error {
	// Open the file
	file, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	// Create a scanner to read the file line by line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// Split the line into SHA sum and filename
		parts := strings.SplitN(line, " ", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid line format: %s", line)
		}

		shasum := parts[0]
		filename := parts[1]
		filename = strings.TrimSpace(filename)

		// Parse the filename
		provider, version, os, arch := parseFilename(filename)

		// // Print the results
		// fmt.Printf("SHA Sum: %s\n", shasum)
		// fmt.Printf("Provider: %s\n", provider)
		// fmt.Printf("Version: %s\n", version)
		// fmt.Printf("OS: %s\n", os)
		// fmt.Printf("Arch: %s\n", arch)
		// fmt.Println()

		// Create the payload struct
		payload := ProviderVersionPayloadPlatform{}
		payload.Data.Type = "registry-provider-version-platforms"
		payload.Data.Attributes.Os = os
		payload.Data.Attributes.Arch = arch
		payload.Data.Attributes.Shasum = shasum
		payload.Data.Attributes.Filename = filename

		// Marshal the struct to JSON
		payloadBytes, err := json.MarshalIndent(payload, "", "  ")
		if err != nil {
			log.Fatal("Error marshaling JSON payload:", err)
		}

		// fmt.Println(string(payloadBytes))
		// Create a request body from the payload
		requestBody := bytes.NewReader(payloadBytes)

		// Construct the URL
		myURL := "https://" + tfeHostname + "/api/v2/organizations/" + organization + "/registry-providers/private/" + organization + "/" + provider + "/versions/" + version + "/platforms"

		// fmt.Println(myURL)

		// Create a new POST request
		req, err := http.NewRequest("POST", myURL, requestBody)
		if err != nil {
			log.Fatal("Error creating request:", err)
		}

		// Set headers
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/vnd.api+json")

		// Send the request
		client := &http.Client{}
		response, err := client.Do(req)
		if err != nil {
			log.Fatal("Error sending request:", err)
		} else {
			fmt.Println("Provider version created")
		}
		defer response.Body.Close()

		// Read and print the response body
		content, err := io.ReadAll(response.Body)
		if err != nil {
			log.Fatal("Error reading response body:", err)
		}

		// fmt.Println("print content", string(content))

		var versionResult map[string]interface{}

		// Unmarshal JSON into a map which is a different approach then knowing the structure. Just as a practice
		err = json.Unmarshal(content, &versionResult)
		if err != nil {
			fmt.Println("error 222: ", err)
			return fmt.Errorf("error: %v", err)
		}
		// fmt.Println("Received JSON:", versionResult)

		// Navigate through the nested structure to find the `shasums-upload` and `shasums-sig-upload`
		var binaryUpload string

		if data, ok := versionResult["data"].(map[string]interface{}); ok {
			if links, ok := data["links"].(map[string]interface{}); ok {
				// Extract shasums-upload
				if upload, ok := links["provider-binary-upload"].(string); ok {
					binaryUpload = upload
				}
			}
		}

		// fmt.Println(binaryUpload)
		// fmt.Println(providerfolder + filename)
		// Helper function to upload a file
		filename = strings.TrimSpace(filename)
		filePath := providerfolder + "/" + filename
		fmt.Println("starting uploading binaries for:", filename)
		uploadFile(binaryUpload,filePath )

	}

	// Check for errors from the scanner
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %v", err)
	}

	return nil
}

// parseFilename extracts provider, version, os, and arch from the filename
func parseFilename(filename string) (provider, version, os, arch string) {
	// Remove the extension
	withoutExt := strings.TrimSuffix(filename, ".zip")

	// Split by underscores
	parts := strings.Split(withoutExt, "_")
	if len(parts) < 4 {
		return "", "", "", ""
	}

	providerFull := strings.Split(parts[0], "-")
	if len(providerFull) < 3 {
		return "", "", "", ""
	}

	provider = providerFull[2]
	version = parts[1]
	os = parts[2]
	arch = parts[3]

	return provider, version, os, arch
}

// uploadFile uploads a file to the specified URL
func uploadFile(url string, filePath string) error {
	// fmt.Println(filePath)
	// fmt.Println(url)

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("error opening file %s: %v", filePath, err)
	}
	defer file.Close()

	// fmt.Println("file working with is", filePath)

	req, err := http.NewRequest("PUT", url, file)
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/octet-stream")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("upload failed: %s", resp.Status)
	}

	return nil
}