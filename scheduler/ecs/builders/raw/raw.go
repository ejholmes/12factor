// Package raw implements a StackBuilder using direct AWS API calls.
package raw

import (
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/remind101/12factor"
	"github.com/remind101/empire/pkg/arn"
)

// DefaultDelimiter is the default delimiter used to delineate between app and
// process in service names.
const DefaultDelimiter = "--"

type ecsClient interface {
	ListServicesPages(*ecs.ListServicesInput, func(*ecs.ListServicesOutput, bool) bool) error
	DeleteService(*ecs.DeleteServiceInput) (*ecs.DeleteServiceOutput, error)
}

// StackBuilder implements the StackBuilder interface for the ECS scheduler.
type StackBuilder struct {
	// ECS Cluster to operate within.
	Cluster string

	// Delimiter to use in the service name to delineate between app and
	// process. The zero value is DefaultDelimiter.
	Delimiter string

	ecs       ecsClient
	splitFunc func(string) (app, process string)
}

func (b *StackBuilder) Build(twelvefactor.App) error {
	// Find ECS services
	// Create ECS services
	// Remove old ECS services
	return nil
}

// Iterates through all of the ECS services for this app and removes them.
func (b *StackBuilder) Remove(app string) error {
	services, err := b.Services(app)
	if err != nil {
		return err
	}

	for _, service := range services {
		if _, err := b.ecs.DeleteService(&ecs.DeleteServiceInput{
			Cluster: aws.String(b.Cluster),
			Service: aws.String(service),
		}); err != nil {
			return err
		}
	}

	return nil
}

// Services iterates through all of the ECS services in this cluster, and
// returns the services that are members of the given app.
func (b *StackBuilder) Services(app string) (map[string]string, error) {
	services := make(map[string]string)

	if err := b.ecs.ListServicesPages(&ecs.ListServicesInput{
		Cluster: aws.String(b.Cluster),
	}, func(resp *ecs.ListServicesOutput, lastPage bool) bool {
		for _, serviceArn := range resp.ServiceArns {
			if serviceArn == nil {
				continue
			}

			id, err := arn.ResourceID(*serviceArn)
			if err != nil {
				return false
			}

			appName, process, ok := b.split(id)
			if !ok {
				continue
			}

			if appName == app {
				services[process] = id
			}
		}

		return true
	}); err != nil {
		return nil, err
	}

	return services, nil
}

func (b *StackBuilder) split(service string) (app, process string, ok bool) {
	parts := strings.SplitN(service, b.delimiter(), 2)
	if len(parts) != 2 {
		return
	}
	app, process, ok = parts[0], parts[1], true
	return
}

func (b *StackBuilder) delimiter() string {
	if b.Delimiter == "" {
		return DefaultDelimiter
	}

	return b.Delimiter
}
