package registry

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
)

const ParamName = "/costgopher/resources"

type Entry struct {
	Service      string `json:"service"`
	ResourceID   string `json:"resourceId"`
	ResourceName string `json:"resourceName"`
	Region       string `json:"region"`
	Creator      string `json:"creator"`
	CreatedAt    string `json:"createdAt"`
}

type Store struct {
	client *ssm.Client
}

func NewStore(cfg aws.Config) *Store {
	return &Store{client: ssm.NewFromConfig(cfg)}
}

func (s *Store) Add(ctx context.Context, entry Entry) error {
	entries, err := s.List(ctx)
	if err != nil {
		return err
	}
	entries = append(entries, entry)
	return s.put(ctx, entries)
}

func (s *Store) List(ctx context.Context) ([]Entry, error) {
	param, err := s.client.GetParameter(ctx, &ssm.GetParameterInput{
		Name: aws.String(ParamName),
	})
	if err != nil {
		var pnf *types.ParameterNotFound
		if errors.As(err, &pnf) {
			return []Entry{}, nil
		}
		return nil, fmt.Errorf("get parameter: %w", err)
	}
	if param.Parameter == nil || param.Parameter.Value == nil {
		return []Entry{}, nil
	}
	var entries []Entry
	if err := json.Unmarshal([]byte(*param.Parameter.Value), &entries); err != nil {
		return nil, fmt.Errorf("unmarshal entries: %w", err)
	}
	return entries, nil
}

func (s *Store) put(ctx context.Context, entries []Entry) error {
	data, err := json.Marshal(entries)
	if err != nil {
		return fmt.Errorf("marshal entries: %w", err)
	}
	_, err = s.client.PutParameter(ctx, &ssm.PutParameterInput{
		Name:      aws.String(ParamName),
		Value:     aws.String(string(data)),
		Type:      types.ParameterTypeString,
		Overwrite: aws.Bool(true),
	})
	if err != nil {
		return fmt.Errorf("put parameter: %w", err)
	}
	return nil
}

func GetWebhookURL(ctx context.Context, cfg aws.Config) (string, error) {
	client := ssm.NewFromConfig(cfg)
	param, err := client.GetParameter(ctx, &ssm.GetParameterInput{
		Name:           aws.String("/costgopher/webhook-url"),
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		return "", fmt.Errorf("get webhook url: %w", err)
	}
	if param.Parameter == nil || param.Parameter.Value == nil {
		return "", fmt.Errorf("webhook URL parameter is empty")
	}
	return *param.Parameter.Value, nil
}
