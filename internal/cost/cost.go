package cost

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer/types"

	"costgopher/internal/format"
)

type Client struct {
	ce *costexplorer.Client
}

func NewClient(cfg aws.Config) *Client {
	return &Client{ce: costexplorer.NewFromConfig(cfg)}
}

func parseAmt(s *string) float64 {
	if s == nil || *s == "" {
		return 0
	}
	v, err := strconv.ParseFloat(*s, 64)
	if err != nil {
		return 0
	}
	return v
}

func fmtCost(f float64) string {
	return fmt.Sprintf("%.2f", f)
}

func GetWeeklyCost(ctx context.Context, ceClient *costexplorer.Client) (total string, services []format.ServiceCost, err error) {
	now := time.Now().UTC()
	end := now.Format("2006-01-02")
	start := now.AddDate(0, 0, -7).Format("2006-01-02")

	result, err := ceClient.GetCostAndUsage(ctx, &costexplorer.GetCostAndUsageInput{
		TimePeriod: &types.DateInterval{
			Start: aws.String(start),
			End:   aws.String(end),
		},
		Granularity: types.GranularityDaily,
		Metrics:     []string{"UnblendedCost"},
		GroupBy: []types.GroupDefinition{
			{Type: types.GroupDefinitionTypeDimension, Key: aws.String("SERVICE")},
		},
	})
	if err != nil {
		return "", nil, fmt.Errorf("GetCostAndUsage: %w", err)
	}

	svcTotals := map[string]float64{}
	var grand float64

	for _, rt := range result.ResultsByTime {
		for _, g := range rt.Groups {
			if len(g.Keys) == 0 {
				continue
			}
			svc := g.Keys[0]
			cost := parseAmt(g.Metrics["UnblendedCost"].Amount)
			svcTotals[svc] += cost
			grand += cost
		}
	}

	for svc, cost := range svcTotals {
		services = append(services, format.ServiceCost{Service: svc, Cost: fmtCost(cost)})
	}
	format.SortServicesByCost(services)
	if services == nil {
		services = []format.ServiceCost{}
	}

	return fmtCost(grand), services, nil
}

func GetMonthToDateCost(ctx context.Context, ceClient *costexplorer.Client) (string, error) {
	now := time.Now().UTC()
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
	end := now.Format("2006-01-02")

	result, err := ceClient.GetCostAndUsage(ctx, &costexplorer.GetCostAndUsageInput{
		TimePeriod: &types.DateInterval{
			Start: aws.String(start),
			End:   aws.String(end),
		},
		Granularity: types.GranularityMonthly,
		Metrics:     []string{"UnblendedCost"},
	})
	if err != nil {
		return "", fmt.Errorf("GetCostAndUsage MTD: %w", err)
	}

	var total float64
	for _, rt := range result.ResultsByTime {
		cost := parseAmt(rt.Total["UnblendedCost"].Amount)
		total += cost
	}
	return fmtCost(total), nil
}

func GetForecast(ctx context.Context, ceClient *costexplorer.Client) (string, error) {
	now := time.Now().UTC()
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
	end := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, time.UTC).Format("2006-01-02")

	result, err := ceClient.GetCostForecast(ctx, &costexplorer.GetCostForecastInput{
		TimePeriod: &types.DateInterval{
			Start: aws.String(start),
			End:   aws.String(end),
		},
		Granularity: types.GranularityMonthly,
		Metric:      types.MetricUnblendedCost,
	})
	if err != nil {
		return "", fmt.Errorf("GetCostForecast: %w", err)
	}

	if result.Total == nil {
		return "0.00", nil
	}
	return fmtCost(parseAmt(result.Total.Amount)), nil
}
