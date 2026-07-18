package main

import (
	"context"
	"log"
	"time"

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
	ceClient = costexplorer.NewFromConfig(awsCfg)
}

func handler(ctx context.Context) error {
	now := time.Now().UTC()
	month := now.Format("January")

	forecastTotal, err := cost.GetForecast(ctx, ceClient)
	if err != nil {
		return err
	}
	log.Printf("Forecast total: $%s", forecastTotal)

	actualSoFar, err := cost.GetMonthToDateCost(ctx, ceClient)
	if err != nil {
		return err
	}
	log.Printf("Actual so far: $%s", actualSoFar)

	msg := format.FormatForecast(forecastTotal, actualSoFar, month)
	log.Printf("Message:\n%s", msg)

	return notifyClient.Send(msg)
}

func main() {
	lambda.Start(handler)
}
