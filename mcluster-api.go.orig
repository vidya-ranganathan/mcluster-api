package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"sync"

	"github.com/gorilla/mux"
)

var clusters = make(map[string]map[string]interface{})
var mu sync.Mutex

func createCluster(clusterName string, workers int) (string, error) {
	mu.Lock()
	defer mu.Unlock()

	// Check if cluster already exists
	if _, exists := clusters[clusterName]; exists {
		return "", fmt.Errorf("Cluster '%s' already exists", clusterName)
	}

	args := []string{"create", "cluster", "--name", clusterName}

	cmd := exec.Command("kind", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("Error creating cluster: %v", err)
	}

	clusters[clusterName] = map[string]interface{}{
		"name": clusterName,
		"node": workers,
		// Add other properties as needed
	}

	return fmt.Sprintf("Cluster '%s' created successfully", clusterName), nil
}

func deleteCluster(clusterName string) (string, error) {
	mu.Lock()
	defer mu.Unlock()

	// Check if the cluster exists
	if _, exists := clusters[clusterName]; !exists {
		return "", fmt.Errorf("Cluster '%s' does not exist", clusterName)
	}

	// Delete the cluster using 'kind delete cluster' command
	args := []string{"delete", "cluster", "--name", clusterName}
	cmd := exec.Command("kind", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("Error deleting cluster '%s': %v", clusterName, err)
	}

	// Remove the cluster from the internal data structure
	delete(clusters, clusterName)

	return fmt.Sprintf("Cluster '%s' deleted successfully", clusterName), nil
}

func abortIfExists(clusterName string, w http.ResponseWriter) {
	if _, exists := clusters[clusterName]; exists {
		abort(w, 409, "cluster already exists")
	}
}

func abortNonExisting(clusterName string, w http.ResponseWriter) {
	if _, exists := clusters[clusterName]; !exists {
		abort(w, 404, "cluster does not exist")
	}
}

func abort(w http.ResponseWriter, code int, message string) {
	w.WriteHeader(code)
	w.Write([]byte(message))
}

type Cluster struct{}

func (c *Cluster) Get(w http.ResponseWriter, r *http.Request) {
	clusterName := mux.Vars(r)["cluster_name"]
	mu.Lock()
	defer mu.Unlock()

	if cluster, exists := clusters[clusterName]; exists {
		json.NewEncoder(w).Encode(cluster)
	} else {
		abort(w, 404, "cluster does not exist")
	}
}

func (c *Cluster) Put(w http.ResponseWriter, r *http.Request) {
	clusterName := mux.Vars(r)["cluster_name"]
	abortIfExists(clusterName, w)

	var args struct {
		Name string `json:"name" binding:"required"`
		Node int    `json:"node" binding:"required"`
		Type int    `json:"type" binding:"required"`
	}

	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		abort(w, 400, "invalid input data")
		return
	}

	// Create the cluster now and insert into the database...
	consoleLog, err := createCluster(clusterName, args.Node)
	if err != nil {
		abort(w, 500, err.Error())
		return
	}

	json.NewEncoder(w).Encode(consoleLog)
}

func (c *Cluster) Delete(w http.ResponseWriter, r *http.Request) {
	clusterName := mux.Vars(r)["cluster_name"]
	response, err := deleteCluster(clusterName)

	if err != nil {
		abort(w, 404, err.Error())
		return
	}

	json.NewEncoder(w).Encode(response)
}

func main() {
	router := mux.NewRouter()
	cluster := &Cluster{}
	router.HandleFunc("/cluster/{cluster_name}", cluster.Get).Methods("GET")
	router.HandleFunc("/cluster/{cluster_name}", cluster.Put).Methods("PUT")
	router.HandleFunc("/cluster/{cluster_name}", cluster.Delete).Methods("DELETE")

	log.Fatal(http.ListenAndServe(":8080", router))
}
