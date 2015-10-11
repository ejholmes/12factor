// Package raw implements a StackBuilder using direct AWS API calls.
package raw

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/remind101/12factor"
	"github.com/remind101/12factor/pkg/aws/arn"
	"github.com/remind101/12factor/pkg/bytesize"
)

// DefaultDelimiter is the default delimiter used to delineate between app and
// process in service names.
const DefaultDelimiter = "--"

type ecsClient interface {
	ListServicesPages(*ecs.ListServicesInput, func(*ecs.ListServicesOutput, bool) bool) error
	DeleteService(*ecs.DeleteServiceInput) (*ecs.DeleteServiceOutput, error)
	RegisterTaskDefinition(*ecs.RegisterTaskDefinitionInput) (*ecs.RegisterTaskDefinitionOutput, error)
	CreateService(*ecs.CreateServiceInput) (*ecs.CreateServiceOutput, error)
}

// StackBuilder implements the StackBuilder interface for the ECS scheduler.
type StackBuilder struct {
	// ECS Cluster to operate within.
	Cluster string

	// Delimiter to use in the service name to delineate between app and
	// process. The zero value is DefaultDelimiter.
	Delimiter string

	// ServiceRole is the name of an IAM role to attach to ECS services that
	// have ELB's attached.
	ServiceRole string

	ecs ecsClient
}

// Build creates or updates ECS services for the app.
func (b *StackBuilder) Build(app twelvefactor.App, processes ...twelvefactor.Process) error {
	for _, process := range processes {
		if err := b.CreateService(app, process); err != nil {
			return err
		}
	}

	return nil
}

// CreateService creates an ECS service for the Process.
func (b *StackBuilder) CreateService(app twelvefactor.App, process twelvefactor.Process) error {
	name := strings.Join([]string{app.ID, process.Name}, b.delimiter())

	taskDefinition, err := b.RegisterTaskDefinition(app, process)
	if err != nil {
		return err
	}

	_, err = b.ecs.CreateService(&ecs.CreateServiceInput{
		Cluster:        aws.String(b.Cluster),
		DesiredCount:   aws.Int64(int64(process.DesiredCount)),
		Role:           aws.String(b.ServiceRole),
		ServiceName:    aws.String(name),
		TaskDefinition: aws.String(taskDefinition),
	})
	return err
}

func (b *StackBuilder) RegisterTaskDefinition(app twelvefactor.App, process twelvefactor.Process) (string, error) {
	family := strings.Join([]string{app.ID, process.Name}, b.delimiter())

	var command []*string
	for _, s := range process.Command {
		ss := s
		command = append(command, &ss)
	}

	var environment []*ecs.KeyValuePair
	for k, v := range twelvefactor.MergeEnv(app.Env, process.Env) {
		environment = append(environment, &ecs.KeyValuePair{
			Name:  aws.String(k),
			Value: aws.String(v),
		})
	}

	resp, err := b.ecs.RegisterTaskDefinition(&ecs.RegisterTaskDefinitionInput{
		Family: aws.String(family),
		ContainerDefinitions: []*ecs.ContainerDefinition{
			{
				Name:        aws.String(process.Name),
				Cpu:         aws.Int64(int64(process.CPUShares)),
				Command:     command,
				Image:       aws.String(app.Image),
				Essential:   aws.Bool(true),
				Memory:      aws.Int64(int64(process.Memory / int(bytesize.MB))),
				Environment: environment,
			},
		},
	})
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s:%d", *resp.TaskDefinition.Family, *resp.TaskDefinition.Revision), nil
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
