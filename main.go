package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
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

// SecretInfo holds the root token and recovery keys
type secretInfo struct {
	RootToken    string
	RecoveryKeys []string
}

func removePeerHandler(event EC2LifecycleHookEvent) error {
	log.Println("removePeerHandler has been initiated...")

	region := os.Getenv("awsRegion")
	secretID := os.Getenv("secretID")

	secretinfo := &secretInfo{}
	svc := secretsmanager.New(session.New(), aws.NewConfig().WithRegion(region))
	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretID),
	}

	output, err := svc.GetSecretValue(input)
	if err != nil {
		log.Println(err)
		return err
	}
	secretString := *output.SecretString
	err = json.Unmarshal([]byte(secretString), secretinfo)
	if err != nil {
		log.Println(err)
		return err
	}

	vaultToken := secretinfo.RootToken

	config := &api.Config{
		Address: os.Getenv("VAULT_ADDR"),
	}
	client, err := api.NewClient(config)
	if err != nil {
		log.Fatal(err)
	}

	client.SetToken(vaultToken)
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
