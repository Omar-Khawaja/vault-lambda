package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/hashicorp/vault/api"
)

// EC2LifecycleHookEventDetail is the field of EC2LifecycleHookEvent that
// contains the information the Vault API will need
type EC2LifecycleHookEventDetail struct {
	EC2InstanceID        string `json:"EC2InstanceId"`
	AutoScalingGroupName string `json:"AutoScalingGroupName"`
	LifecycleActionToken string `json:"LifecycleActionToken"`
	LifecycleHookName    string `json:"LifecycleHookName"`
	NotificationMetadata string `json:"NotificationMetadata"`
}

// EC2LifecycleHookEvent is the event that removePeerHandler is called with
type EC2LifecycleHookEvent struct {
	Detail EC2LifecycleHookEventDetail `json:"detail"`
}

func removePeerHandler(event EC2LifecycleHookEvent) error {
	log.Println("removePeerHandler has been initiated...")
	config := &api.Config{
		Address: os.Getenv("VAULT_ADDR"),
	}
	client, err := api.NewClient(config)
	if err != nil {
		log.Fatal(err)
	}
	client.SetToken(os.Getenv("VAULT_TOKEN"))
	logical := client.Logical()

	// The Terraform that spins up the Vault cluster configures the Vault
	// node IDs (which are needed for peer removal) to be the same as the
	// EC2 instance IDs
	vaultNodeID := event.Detail.EC2InstanceID
	log.Printf("Vault node ID to be removed: %s\n", vaultNodeID)

	_, err = logical.Write("sys/storage/raft/remove-peer",
		map[string]interface{}{
			"server_id": vaultNodeID,
		})
	if err != nil {
		log.Println(err)
		return err
	}

	fmt.Printf("%s has been purged from raft peer list\n", vaultNodeID)
	log.Println("removePeerHandler is now finished")
	return nil
}

func main() {
	lambda.Start(removePeerHandler)
}
