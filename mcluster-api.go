package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"

	"github.com/gorilla/mux"
)

var clusters = make(map[string]map[string]interface{})
var clusterID = 0

func createCluster(clusterName string, workers int) (string, error) {
	args := []string{"create", "cluster", "--name", clusterName}

	cmd := exec.Command("kind", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("Error creating cluster: %v", err)
	}

	return fmt.Sprintf("Cluster '%s' created successfully", clusterName), nil
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

	workerNodeCount := args.Node
	for workerNodeCount > 0 {
		workerNodeCount--
		fmt.Println(workerNodeCount)
	}

	// Create the cluster now and insert into the database...
	consoleLog, err := createCluster(clusterName, args.Node)
	if err != nil {
		abort(w, 500, err.Error())
		return
	}

	clusterID++
	clusters[strconv.Itoa(clusterID)] = map[string]interface{}{
		"name": args.Name,
		"node": args.Node,
		"type": args.Type,
	}
	fmt.Println(clusters)
	json.NewEncoder(w).Encode(consoleLog)
}

func (c *Cluster) Delete(w http.ResponseWriter, r *http.Request) {
	clusterName := mux.Vars(r)["cluster_name"]
	abortNonExisting(clusterName, w)

	for key, cluster := range clusters {
		if cluster["name"].(string) == clusterName {
			delete(clusters, key)
			break
		}
	}
	w.WriteHeader(http.StatusNoContent)
}

func main() {
	router := mux.NewRouter()
	cluster := &Cluster{}
	router.HandleFunc("/cluster/{cluster_name}", cluster.Get).Methods("GET")
	router.HandleFunc("/cluster/{cluster_name}", cluster.Put).Methods("PUT")
	router.HandleFunc("/cluster/{cluster_name}", cluster.Delete).Methods("DELETE")

	log.Fatal(http.ListenAndServe(":8080", router))
}
