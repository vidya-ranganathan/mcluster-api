package main

import (
	"fmt"
	"net/http"
	"os"
)

func DeleteVerb(url string) error {
	// Create a simple DELETE request
	req, err := http.NewRequest("DELETE", url, nil)
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

	fmt.Println("Response Status:", resp.Status)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Failed to delete cluster, response status: %s", resp.Status)
	}

	// Optionally, you can add more logic here to handle the response body if needed
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

	// Build the URL for the DELETE request
	deleteUrl := fmt.Sprintf("http://%s:%s/cluster/%s", serverHost, serverPort, clusterName)

	// Call the DELETE function to delete a cluster
	err := DeleteVerb(deleteUrl)
	if err != nil {
		fmt.Println("Error:", err)
	}
}
