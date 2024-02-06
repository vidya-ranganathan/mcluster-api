package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func delete(url string, payload map[string]interface{}) error {
	// Convert payload to JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("Error marshaling JSON: %v", err)
	}

	// Create a request with the payload
	req, err := http.NewRequest("DELETE", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("Error creating request: %v", err)
	}

	// Set the request headers
	req.Header.Set("Content-Type", "application/json")

	// Make the DELETE request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Error making request: %v", err)
	}
	defer resp.Body.Close()

	// Print the response status and body
	fmt.Println("Response Status:", resp.Status)
	// Read and print the response body if needed
	// responseBody, _ := ioutil.ReadAll(resp.Body)
	// fmt.Println("Response Body:", string(responseBody))

	return nil
}

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: go run main.go <server_host> <server_port> <cluster_name>")
		os.Exit(1)
	}

	// Parse command-line arguments
	serverHost := os.Args[1]
	serverPort := os.Args[2]
	clusterName := os.Args[3]

	// Build the URL
	url := fmt.Sprintf("http://%s:%s/cluster/%s", serverHost, serverPort, clusterName)

	// Define the payload to be sent in the request body
	payload := map[string]interface{}{
		"name": "your_cluster_name",
		"node": 3, // Replace with the desired number of nodes
		"type": 1, // Replace with the desired type
	}

	// Call the DELETE function
	err := delete(url, payload)
	if err != nil {
		fmt.Println("Error:", err)
	}
}
