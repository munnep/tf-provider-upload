package gpg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"bufio"
)

type Payload struct {
	Data struct {
		Type       string `json:"type"`
		Attributes struct {
			Namespace  string `json:"namespace"`
			AsciiArmor string `json:"ascii-armor"`
		} `json:"attributes"`
	} `json:"data"`
}

type PayloadResponse struct {
	Data struct {
		Attributes struct {
			KeyID string `json:"key-id"`
		} `json:"attributes"`
	} `json:"data"`
}

type PayloadResponseCheckGPG struct {
	Data []struct {
		Attributes struct {
			ASCIIArmor string `json:"ascii-armor"`
			KeyID      string `json:"key-id"`
		} `json:"attributes"`
	} `json:"data"`
}

func CheckGPGKey(gpgPublicKeyString string, token string, tfeHostname string, organization string) (bool, string) {

	// Construct the URL
	myURL := fmt.Sprintf("https://%s/api/registry/private/v2/gpg-keys?filter[namespace]=%s", tfeHostname, organization)

	// Create a new GET request
	req, err := http.NewRequest("GET", myURL, nil)
	if err != nil {
		fmt.Errorf("failed to create request: %v", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/vnd.api+json")

	// Create an HTTP client and make the request
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		fmt.Errorf("request failed: %v", err)
	}
	defer res.Body.Close()

	// Read the response body
	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Errorf("failed to read response body: %v", err)
	}

	// Check for non-success status codes
	if res.StatusCode > 299 {
		fmt.Errorf("response failed with status code: %d and body: %s", res.StatusCode, body)
	}

	// Unmarshal the JSON response into the struct
	var payload PayloadResponseCheckGPG
	if err := json.Unmarshal(body, &payload); err != nil {
		fmt.Errorf("error unmarshaling JSON response: %v", err)
	}

	// Collect KeyIDs from the response
	for _, item := range payload.Data {
		// fmt.Println(item.Attributes.KeyID)
		if gpgPublicKeyString == item.Attributes.ASCIIArmor {
			// the GPG public is already uploaded
			return true, item.Attributes.KeyID
		}
	}

	// Return the key-ids
	return false, "unknown"
}


func UploadGPGKey(gpgPublicKeyString string, token string, tfeHostname string, organization string) string {

	// fmt.Println(gpgPublicKeyString)
	// Create the payload struct
	payload := Payload{}
	payload.Data.Type = "gpg-keys"
	payload.Data.Attributes.Namespace = organization
	payload.Data.Attributes.AsciiArmor = gpgPublicKeyString

	// Marshal the struct to JSON
	payloadBytes, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		log.Fatal("Error marshaling JSON payload:", err)
	}

	// fmt.Println(string(payloadBytes))
	// Create a request body from the payload
	requestBody := bytes.NewReader(payloadBytes)

	// Construct the URL
	myURL := "https://" + tfeHostname + "/api/registry/private/v2/gpg-keys"

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
	}
	defer response.Body.Close()

	// Read and print the response body
	content, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatal("Error reading response body:", err)
	}

	// Unmarshal the JSON response into the struct
	var responsePayload PayloadResponse
	err = json.Unmarshal(content, &responsePayload)
	if err != nil {
		log.Fatal("Error unmarshaling JSON response:", err)
	}

	// Extract the key-id
	gpgKeyID := responsePayload.Data.Attributes.KeyID

	// Return the key-id
	return gpgKeyID

}

func Keys(providerfolder string, gpg_key string) string {
	check_gpg_tool()
	// create_key()             // going to use the keystring we already have.....figure this out later
	keystring := which_key_to_use(gpg_key)
	generate_public_gpg_key_file(keystring)
	check_shasum_tool()
	generate_shasum_files(keystring, providerfolder)
	generate_shasum_sig_file(keystring, providerfolder)

	gpg_public_key_as_string := gpg_public_key_as_string()
	return gpg_public_key_as_string
}

func check_gpg_tool() {
	fmt.Println("Check the gpg tool is installed")

	cmd := exec.Command("gpg", "--version")
	err := cmd.Run()

	if err != nil {
		fmt.Println("GPG is not installed or not found in PATH.")
	} else {
		fmt.Println("GPG is installed.")
	}
}

func check_shasum_tool() {
	fmt.Println("Check the shasum tool is installed")

	cmd := exec.Command("shasum", "--version")
	err := cmd.Run()

	if err != nil {
		fmt.Println("shasum is not installed or not found in PATH.")
	} else {
		fmt.Println("shasum is installed.")
	}
}

func create_key() {
	// Create a batch file for GPG key generation
	batchFile := ".tf-provider-upload/gen-key-batch.txt"
	batchContent := `
	Key-Type: 1
	Key-Length: 4096
	Subkey-Type: 1
	Subkey-Length: 4096
	Name-Real: tf-provider-upload
	Name-Email: tf-provider-upload@example.com
	Expire-Date: 0
	%no-protection
	%commit
	`
	// Write batch content to a file
	err := os.WriteFile(batchFile, []byte(batchContent), 0600)
	if err != nil {
		fmt.Println("Error creating batch file:", err)
		return
	}

	// Run the GPG command with the batch file
	cmd := exec.Command("gpg", "--batch", "--gen-key", batchFile)
	// cmd.Stdout = os.Stdout
	// cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		fmt.Println("Error generating GPG key:", err)
		return
	}
	fmt.Println("GPG key generated successfully.")

}

func which_key_to_use(gpg_key string) string {
	var keystring string

	if gpg_key != "" {
		return gpg_key
	} else {
		fmt.Print("Enter your gpg key string: ")
		fmt.Scanln(&keystring)
		// fmt.Println("keystring:", keystring )
		return keystring
	}
}

func generate_shasum_files(keystring string, providerfolder string) {
	// Construct the command string
	cmdStr := fmt.Sprintf("shasum -a 256 *.zip > files.SHA256SUMS")

	// Use the shell to execute the command
	cmd := exec.Command("sh", "-c", cmdStr)

	cmd.Dir = providerfolder
	// Run the command
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error generating SHA256SUM file:", err)
		return
	}

	fmt.Println("SHA256SUM file generated successfully.")

}

func generate_public_gpg_key_file(keystring string) {
	// remove the currently existing gpg-key.pub file
	gpgKeyPublic := ".tf-provider-upload/gpg-key.pub"
	err := os.Remove(gpgKeyPublic)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("gpgKeyPublic file does not exist")
		} else {
			fmt.Printf("Error removing file: .tf-provider-upload/gpg-key.pub: %v", err)
		}
	} else {
		fmt.Println("gpgKeyPublic file removed successfully")
	}

	// Construct the command string
	cmdStr := fmt.Sprintf("gpg -o .tf-provider-upload/gpg-key.pub -a --export %s", keystring)

	// Use the shell to execute the command
	cmd := exec.Command("sh", "-c", cmdStr)

	// Run the command
	err = cmd.Run()
	if err != nil {
		fmt.Println("Error creating gpg public key:", err)
		return
	}

	fmt.Println("gpg public key created successfully")

}

func generate_shasum_sig_file(keystring string, providerfolder string) {

	// remove the currently existing signature file
	signatureFile := fmt.Sprintf(providerfolder + "/files.SHA256SUMS.sig")
	err := os.Remove(signatureFile)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Signature file does not exist")
		} else {
			fmt.Printf("Error removing file: %v\n", err)
		}
	} else {
		fmt.Println("Signature file removed successfully")
	}

	// Construct the command string
	cmdStr := fmt.Sprintf("gpg  --default-key %s -sb %s/files.SHA256SUMS", keystring, providerfolder)

	// Use the shell to execute the command
	cmd := exec.Command("sh", "-c", cmdStr)

	// Run the command
	err = cmd.Run()
	if err != nil {
		fmt.Println("Error generating SHA256SUM signature file:", err)
		return
	}

	fmt.Println("SHA256SUM signature file generated successfully.")

}

func gpg_public_key_as_string() string {
	cmd := exec.Command("sh", "-c", "sed 's/$/\\\\n/g' .tf-provider-upload/gpg-key.pub | tr -d '\n\r'")

	// Create a buffer to capture the output
	var out bytes.Buffer
	cmd.Stdout = &out

	// Run the command
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error running command:", err)
		return "error"
	}

	originalString := out.String()
	// remove some extra characters
	result := strings.ReplaceAll(originalString, "\\n", "\n")
	// Convert the output buffer to a string

	// Print the result
	// fmt.Println("Output:", result)
	return result
}






//
//
//
//



// UploadProviderVersionPlatform is the main function to process the files
func GetProviderNameVersion(providerfolder string) (string, string, error) {
	// Construct the path to the SHA256SUMS file
	filename := providerfolder + "/files.SHA256SUMS"

    // Process the file and get the provider and version
    provider, version, err := processFile(filename)
    if err != nil {
        return "", "", fmt.Errorf("error processing file: %v", err)
    }

    return provider, version, nil
	
}

// processFile reads the file and parses each line
func processFile(filepath string) (string, string, error) {
    // Open the file
    file, err := os.Open(filepath)
    if err != nil {
        return "", "", fmt.Errorf("error opening file: %v", err)
    }
    defer file.Close()

    // Create a scanner to read the file line by line
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        line := scanner.Text()

        // Split the line into SHA sum and filename
        parts := strings.SplitN(line, " ", 2)
        if len(parts) != 2 {
            return "", "", fmt.Errorf("invalid line format: %s", line)
        }

        // filename := parts[1]

        // Parse the filename
        provider, version := parseFilename(parts[1])

        // Return the first provider and version found
        return provider, version, nil
    }

    // Check for errors from the scanner
    if err := scanner.Err(); err != nil {
        return "", "", fmt.Errorf("error reading file: %v", err)
    }

    return "", "", fmt.Errorf("no valid lines found in file")
}






// parseFilename extracts provider, version, os, and arch from the filename
func parseFilename(filename string) (provider, version string) {
	// Remove the extension
	withoutExt := strings.TrimSuffix(filename, ".zip")
	
	// Split by underscores
	parts := strings.Split(withoutExt, "_")
	if len(parts) < 4 {
		return "", "" 
	}
	
	providerFull := strings.Split(parts[0], "-")
	if len(providerFull) < 3 {
		return "", "" 
	}

	provider = providerFull[2]
	version = parts[1]
	// os = parts[2]
	// arch = parts[3]
	
	// return provider, version, os, arch
	return provider, version
}

