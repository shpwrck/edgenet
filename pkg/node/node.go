package node

import (
	"encoding/json"
	"fmt"
	"time"

	"headnode/pkg/authorization"
	"headnode/pkg/node/infrastructure"

	namecheap "github.com/billputer/go-namecheap"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SetHostname generates token to be used on adding a node onto the cluster
func SetHostname(hostRecord namecheap.DomainDNSHost) (bool, string) {
	client, err := authorization.CreateNamecheapClient()
	if err != nil {
		return false, "Unknown"
	}
	result, state := infrastructure.SetHostname(client, hostRecord)
	return result, state
}

// CreateJoinToken generates token to be used on adding a node onto the cluster
func CreateJoinToken(ttl string, hostname string) string {
	clientset, err := authorization.CreateClientSet()
	if err != nil {
		panic(err.Error())
	}
	duration, _ := time.ParseDuration(ttl)
	token, err := infrastructure.CreateToken(clientset, duration, hostname)
	if err != nil {
		return "error"
	}
	return token
}

// GetList uses clientset to get node list of the cluster
func GetList() []string {
	clientset, err := authorization.CreateClientSet()
	if err != nil {
		panic(err.Error())
	}

	nodesRaw, err := clientset.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	nodes := make([]string, len(nodesRaw.Items))
	for i, nodeRow := range nodesRaw.Items {
		nodes[i] = nodeRow.Name
	}

	return nodes
}

// GetList uses clientset to get node list of the cluster that contains Ready State info
func GetStatusList() []byte {
	type nodeStatus struct {
		Node  string `json:"node"`
		Ready string `json:"ready"`
	}
	clientset, err := authorization.CreateClientSet()
	if err != nil {
		panic(err.Error())
	}

	nodesRaw, err := clientset.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	nodesArr := make([]nodeStatus, len(nodesRaw.Items))
	for i, nodeRow := range nodesRaw.Items {
		nodesArr[i].Node = nodeRow.Name
		for _, conditionRow := range nodeRow.Status.Conditions {
			if conditionType := conditionRow.Type; conditionType == "Ready" {
				nodesArr[i].Ready = string(conditionRow.Status)
			}
		}
	}
	nodesJSON, _ := json.Marshal(nodesArr)

	return nodesJSON
}

// getNodeByHostname uses clientset to get namespace requested
func getNodeByHostname(hostname string) (string, error) {
	clientset, err := authorization.CreateClientSet()
	if err != nil {
		panic(err.Error())
	}

	// Examples for error handling:
	// - Use helper functions like e.g. errors.IsNotFound()
	// - And/or cast to StatusError and use its properties like e.g. ErrStatus.Message
	_, err = clientset.CoreV1().Nodes().Get(hostname, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		fmt.Printf("Node %s not found\n", hostname)
		return "Node not found", err
	} else if statusError, isStatus := err.(*errors.StatusError); isStatus {
		fmt.Printf("Error getting node %s: %v\n",
			hostname, statusError.ErrStatus)
		return "Error getting node", err
	} else if err != nil {
		panic(err.Error())
	} else {
		return "", nil
	}
}
