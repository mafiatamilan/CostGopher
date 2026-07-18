package main

import (
	"context"
	"log"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"

	"costgopher/internal/format"
	"costgopher/internal/notify"
	"costgopher/internal/registry"
)

type EventBridgeEvent struct {
	Version    string                 `json:"version"`
	ID         string                 `json:"id"`
	DetailType string                 `json:"detail-type"`
	Source     string                 `json:"source"`
	Account    string                 `json:"account"`
	Time       string                 `json:"time"`
	Region     string                 `json:"region"`
	Resources  []string               `json:"resources"`
	Detail     map[string]interface{} `json:"detail"`
}

var (
	notifyClient *notify.Client
	regStore     *registry.Store
	awsCfg       aws.Config
)

func init() {
	ctx := context.Background()
	var err error
	awsCfg, err = config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	webhookURL, err := registry.GetWebhookURL(ctx, awsCfg)
	if err != nil {
		log.Fatalf("webhook: %v", err)
	}

	notifyClient = notify.NewClient(webhookURL)
	regStore = registry.NewStore(awsCfg)
}

func handler(ctx context.Context, event EventBridgeEvent) error {
	detail := event.Detail

	info := format.ExtractResourceInfo(detail, event.Time)

	msg := format.FormatAlert(info)
	log.Printf("Alert: %s", msg)

	if err := notifyClient.Send(msg); err != nil {
		return err
	}

	if info.ResourceID != "" {
		entry := registry.Entry{
			Service:      info.Service,
			ResourceID:   info.ResourceID,
			ResourceName: info.ResourceName,
			Region:       info.Region,
			Creator:      info.Creator,
			CreatedAt:    event.Time,
		}
		log.Printf("Registry entry: service=%s id=%s name=%s", entry.Service, entry.ResourceID, entry.ResourceName)

		if err := regStore.Add(ctx, entry); err != nil {
			log.Printf("registry add error: %v", err)
		}
	}

	return nil
}

func main() {
	lambda.Start(handler)
}
