package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"sync"

	"github.com/gorilla/mux"
)

var clusters = make(map[string]string) // Maps clusterName to clusterID for simplicity
var mu sync.Mutex

// createCluster generates a cluster ID based on the cluster name using SHA-256
// and stores the cluster name and its ID in a map.
// It also invokes the kind command to create a cluster.
func createCluster(clusterName string) (string, error) {
	mu.Lock()

	if _, exists := clusters[clusterName]; exists {
		return "", fmt.Errorf("Cluster '%s' already exists", clusterName)
	}

	// Generating SHA-256 hash of the clusterName to use as clusterID
	hasher := sha256.New()
	hasher.Write([]byte(clusterName))
	clusterID := hex.EncodeToString(hasher.Sum(nil))

	clusters[clusterName] = clusterID

	mu.Unlock()

	// Execute the kind command to create the cluster
	args := []string{"create", "cluster", "--name", clusterName}
	cmd := exec.Command("kind", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("Error creating cluster: %v", err)
	}

	return clusterID, nil
}

// deleteCluster removes the specified cluster by invoking the kind command.
// It also deletes the cluster from the internal map.
func deleteCluster(clusterName string) (string, error) {
	mu.Lock()

	if _, exists := clusters[clusterName]; !exists {
		return "", fmt.Errorf("Cluster '%s' does not exist", clusterName)
	}
	delete(clusters, clusterName)

	mu.Unlock()

	// Execute the kind command to delete the cluster
	args := []string{"delete", "cluster", "--name", clusterName}
	cmd := exec.Command("kind", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("Error deleting cluster '%s': %v", clusterName, err)
	}

	return fmt.Sprintf("Cluster '%s' deleted successfully", clusterName), nil
}

func abort(w http.ResponseWriter, code int, message string) {
	w.WriteHeader(code)
	w.Write([]byte(message))
}

type Cluster struct{}

func (c *Cluster) Get(w http.ResponseWriter, r *http.Request) {
	clusterName := mux.Vars(r)["cluster_name"]
	mu.Lock()
	clusterID, exists := clusters[clusterName]
	mu.Unlock()

	if exists {
		json.NewEncoder(w).Encode(map[string]string{"name": clusterName, "clusterID": clusterID})
	} else {
		abort(w, 404, "cluster does not exist")
	}
}

func (c *Cluster) Put(w http.ResponseWriter, r *http.Request) {
	clusterName := mux.Vars(r)["cluster_name"]

	mu.Lock()
	if _, exists := clusters[clusterName]; exists {
		mu.Unlock()
		abort(w, 409, "cluster already exists")
		return
	}
	mu.Unlock()

	clusterID, err := createCluster(clusterName)
	if err != nil {
		abort(w, 500, err.Error())
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"name": clusterName, "clusterID": clusterID})
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
