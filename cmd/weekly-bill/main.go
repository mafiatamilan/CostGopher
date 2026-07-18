package main

import (
	"context"
	"log"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer"

	"costgopher/internal/cost"
	"costgopher/internal/format"
	"costgopher/internal/notify"
	"costgopher/internal/registry"
)

var (
	notifyClient *notify.Client
	regStore     *registry.Store
	ceClient     *costexplorer.Client
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
	ceClient = costexplorer.NewFromConfig(awsCfg)
}

func handler(ctx context.Context) error {
	total, services, err := cost.GetWeeklyCost(ctx, ceClient)
	if err != nil {
		return err
	}
	log.Printf("Weekly total: $%s, services: %d", total, len(services))

	resources, err := regStore.List(ctx)
	if err != nil {
		log.Printf("registry list error: %v", err)
		resources = []registry.Entry{}
	}

	res := make([]format.RegistryEntry, len(resources))
	for i, r := range resources {
		res[i] = format.RegistryEntry{
			Service:      r.Service,
			ResourceID:   r.ResourceID,
			ResourceName: r.ResourceName,
			Region:       r.Region,
			Creator:      r.Creator,
			CreatedAt:    r.CreatedAt,
		}
	}

	msg := format.FormatWeeklyBill(total, services, res)
	log.Printf("Message:\n%s", msg)

	return notifyClient.Send(msg)
}

func main() {
	lambda.Start(handler)
}
