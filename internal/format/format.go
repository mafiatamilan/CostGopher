package format

import (
	"fmt"
	"sort"
	"strings"
)

var serviceNames = map[string]string{
	"ec2.amazonaws.com":                    "EC2",
	"rds.amazonaws.com":                    "RDS",
	"s3.amazonaws.com":                     "S3",
	"lambda.amazonaws.com":                 "Lambda",
	"elasticache.amazonaws.com":            "ElastiCache",
	"es.amazonaws.com":                     "OpenSearch",
	"redshift.amazonaws.com":               "Redshift",
	"kinesis.amazonaws.com":                "Kinesis",
	"dynamodb.amazonaws.com":               "DynamoDB",
	"sqs.amazonaws.com":                    "SQS",
	"sns.amazonaws.com":                    "SNS",
	"ecs.amazonaws.com":                    "ECS",
	"eks.amazonaws.com":                    "EKS",
	"autoscaling.amazonaws.com":            "AutoScaling",
	"elasticloadbalancing.amazonaws.com":   "ELB",
}

var eventActions = map[string]string{
	"RunInstances":             "launched a new server",
	"CreateInstance":           "launched a new server",
	"CreateDBInstance":         "created a new database",
	"CreateBucket":             "created a new storage bucket",
	"CreateFunction":           "created a new serverless function",
	"CreateFunction20150331":   "created a new serverless function",
	"CreateCacheCluster":       "created a new cache cluster",
	"CreateElasticsearchDomain": "created a new search domain",
	"CreateCluster":            "created a new cluster",
	"CreateStream":             "created a new data stream",
	"CreateTable":              "created a new table",
	"CreateQueue":              "created a new queue",
	"CreateTopic":              "created a new notification topic",
	"CreateService":            "created a new service",
	"RunTask":                  "started a new task",
	"CreateAutoScalingGroup":   "created a new auto scaling group",
	"CreateLoadBalancer":       "created a new load balancer",
}

func friendlyService(eventSource string) string {
	if s, ok := serviceNames[eventSource]; ok {
		return s
	}
	parts := strings.SplitN(eventSource, ".", 2)
	if len(parts) > 0 {
		return strings.Title(parts[0])
	}
	return eventSource
}

func friendlyAction(eventName string) string {
	if a, ok := eventActions[eventName]; ok {
		return a
	}
	return fmt.Sprintf("executed %s", eventName)
}

type ResourceInfo struct {
	Service      string
	Action       string
	ResourceID   string
	ResourceName string
	DetailLine   string
	Creator      string
	CreatorType  string
	Region       string
	EventTime    string
}

type ServiceCost struct {
	Service string
	Cost    string
}

type RegistryEntry struct {
	Service      string `json:"service"`
	ResourceID   string `json:"resourceId"`
	ResourceName string `json:"resourceName"`
	Region       string `json:"region"`
	Creator      string `json:"creator"`
	CreatedAt    string `json:"createdAt"`
}

func val(m map[string]interface{}, key string) string {
	v, ok := m[key]
	if !ok {
		return ""
	}
	s, _ := v.(string)
	return s
}

func sub(m map[string]interface{}, key string) map[string]interface{} {
	v, ok := m[key]
	if !ok {
		return nil
	}
	r, _ := v.(map[string]interface{})
	return r
}

func arr(m map[string]interface{}, key string) []interface{} {
	v, ok := m[key]
	if !ok {
		return nil
	}
	r, _ := v.([]interface{})
	return r
}

func findNameTag(data map[string]interface{}) string {
	if tss := sub(data, "tagSpecificationSet"); tss != nil {
		for _, item := range arr(tss, "items") {
			if m, ok := item.(map[string]interface{}); ok {
				for _, tag := range arr(m, "tags") {
					if t, ok := tag.(map[string]interface{}); ok {
						if strings.EqualFold(val(t, "key"), "Name") {
							return val(t, "value")
						}
					}
				}
			}
		}
	}
	if tags := arr(data, "tags"); tags != nil {
		for _, tag := range tags {
			if t, ok := tag.(map[string]interface{}); ok {
				if strings.EqualFold(val(t, "key"), "Name") {
					return val(t, "value")
				}
			}
		}
	}
	return ""
}

func extractResourceID(detail map[string]interface{}, eventSource string) string {
	re := sub(detail, "responseElements")
	if re == nil {
		return ""
	}
	switch eventSource {
	case "ec2.amazonaws.com":
		if is := sub(re, "instancesSet"); is != nil {
			if items := arr(is, "items"); len(items) > 0 {
				if m, ok := items[0].(map[string]interface{}); ok {
					return val(m, "instanceId")
				}
			}
		}
	case "rds.amazonaws.com":
		if id := val(re, "dBInstanceIdentifier"); id != "" {
			return id
		}
		if arn := val(re, "dBInstanceArn"); arn != "" {
			parts := strings.Split(arn, ":")
			return parts[len(parts)-1]
		}
	case "lambda.amazonaws.com":
		return val(re, "functionName")
	case "elasticache.amazonaws.com":
		return val(re, "cacheClusterId")
	case "es.amazonaws.com":
		return val(re, "domainName")
	case "dynamodb.amazonaws.com":
		if td := sub(re, "tableDescription"); td != nil {
			return val(td, "tableName")
		}
	case "redshift.amazonaws.com":
		return val(re, "clusterIdentifier")
	case "kinesis.amazonaws.com":
		return val(re, "streamName")
	case "sqs.amazonaws.com":
		if url := val(re, "queueUrl"); url != "" {
			parts := strings.Split(url, "/")
			return parts[len(parts)-1]
		}
	case "autoscaling.amazonaws.com":
		return val(re, "autoScalingGroupName")
	case "elasticloadbalancing.amazonaws.com":
		return val(re, "loadBalancerName")
	}
	return ""
}

func extractResourceIDFromRequest(detail map[string]interface{}, eventSource string) string {
	rp := sub(detail, "requestParameters")
	if rp == nil {
		return ""
	}
	switch eventSource {
	case "s3.amazonaws.com":
		return val(rp, "bucketName")
	case "sns.amazonaws.com":
		return val(rp, "name")
	case "sqs.amazonaws.com":
		return val(rp, "queueName")
	case "ecs.amazonaws.com":
		return val(rp, "serviceName")
	}
	return ""
}

func extractDetailLine(detail map[string]interface{}, eventSource string) string {
	rp := sub(detail, "requestParameters")
	if rp == nil {
		return ""
	}
	switch eventSource {
	case "ec2.amazonaws.com":
		if is := sub(rp, "instancesSet"); is != nil {
			if items := arr(is, "items"); len(items) > 0 {
				if m, ok := items[0].(map[string]interface{}); ok {
					if t := val(m, "instanceType"); t != "" {
						return fmt.Sprintf("EC2 server (%s)", t)
					}
				}
			}
		}
	case "rds.amazonaws.com":
		if t := val(rp, "dBInstanceClass"); t != "" {
			return fmt.Sprintf("RDS database (%s)", t)
		}
		return "RDS database"
	}
	return ""
}

func ExtractResourceInfo(detail map[string]interface{}, eventTime string) *ResourceInfo {
	eventSource := val(detail, "eventSource")
	eventName := val(detail, "eventName")
	region := val(detail, "awsRegion")

	service := friendlyService(eventSource)
	action := friendlyAction(eventName)

	uid := sub(detail, "userIdentity")
	creator := ""
	creatorType := "IAM user"
	if uid != nil {
		creator = val(uid, "userName")
		if creator == "" {
			creator = val(uid, "arn")
		}
		if val(uid, "type") == "AssumedRole" || strings.Contains(creator, ":assumed-role/") {
			creatorType = "IAM role"
		}
	}
	if creator == "" {
		creator = "unknown"
	}

	rp := sub(detail, "requestParameters")
	resourceName := ""
	if rp != nil {
		resourceName = findNameTag(rp)
	}

	resourceID := extractResourceID(detail, eventSource)
	if resourceID == "" {
		resourceID = extractResourceIDFromRequest(detail, eventSource)
	}

	detailLine := extractDetailLine(detail, eventSource)
	if detailLine == "" {
		detailLine = service
	}

	return &ResourceInfo{
		Service:      service,
		Action:       action,
		ResourceID:   resourceID,
		ResourceName: resourceName,
		DetailLine:   detailLine,
		Creator:      creator,
		CreatorType:  creatorType,
		Region:       region,
		EventTime:    eventTime,
	}
}

func FormatAlert(info *ResourceInfo) string {
	var b strings.Builder
	b.WriteString("🟢 New AWS resource created\n")
	b.WriteString(fmt.Sprintf("Who: %s (%s)\n", info.Creator, info.CreatorType))

	what := info.Action
	if info.DetailLine != "" {
		what = fmt.Sprintf("%s — %s", info.Action, info.DetailLine)
	}
	if info.Region != "" {
		what = fmt.Sprintf("%s in %s", what, info.Region)
	}
	b.WriteString(fmt.Sprintf("What: %s\n", what))

	if info.ResourceID != "" {
		if info.ResourceName != "" {
			b.WriteString(fmt.Sprintf("Name: %s (%s)\n", info.ResourceName, info.ResourceID))
		} else {
			b.WriteString(fmt.Sprintf("Name: %s\n", info.ResourceID))
		}
	}
	if info.EventTime != "" {
		b.WriteString(fmt.Sprintf("When: %s\n", info.EventTime))
	}
	b.WriteString("⚠️  This service has a cost — keep an eye on it.")
	return b.String()
}

func FormatWeeklyBill(total string, services []ServiceCost, resources []RegistryEntry) string {
	var b strings.Builder
	b.WriteString("📊 Weekly AWS Bill Update\n")
	b.WriteString(fmt.Sprintf("Total this week: $%s\n\n", total))

	resByService := map[string][]RegistryEntry{}
	for _, r := range resources {
		resByService[r.Service] = append(resByService[r.Service], r)
	}

	for i, svc := range services {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString(fmt.Sprintf("%s — $%s", svc.Service, svc.Cost))
		for _, r := range resByService[svc.Service] {
			if r.ResourceName != "" {
				b.WriteString(fmt.Sprintf("\n  • %s (%s)", r.ResourceName, r.ResourceID))
			} else {
				b.WriteString(fmt.Sprintf("\n  • %s", r.ResourceID))
			}
		}
	}
	return b.String()
}

func FormatForecast(forecastTotal, actualSoFar, month string) string {
	return fmt.Sprintf("🔮 Mid-Month Forecast (as of the 15th)\nProjected total for %s: $%s\nSo far spent: $%s", month, forecastTotal, actualSoFar)
}

func SortServicesByCost(services []ServiceCost) {
	sort.Slice(services, func(i, j int) bool {
		return services[i].Cost > services[j].Cost
	})
}
