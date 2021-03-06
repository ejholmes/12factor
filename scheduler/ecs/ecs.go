// Package ecs provides a scheduler for running 12factor applications using ECS.
package ecs

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/remind101/12factor"
	"github.com/remind101/12factor/pkg/aws/arn"
	"github.com/remind101/12factor/scheduler/ecs/builders/raw"
)

// ProcessNotFoundError is returned when attempting to operate on a process that
// does not exist.
type ProcessNotFoundError struct {
	Process string
}

// Error implements the error interface.
func (e *ProcessNotFoundError) Error() string {
	return fmt.Sprintf("%s process not found", e.Process)
}

// ecsClient represents a client for interacting with ECS.
type ecsClient interface {
	UpdateService(*ecs.UpdateServiceInput) (*ecs.UpdateServiceOutput, error)
	ListTasks(*ecs.ListTasksInput) (*ecs.ListTasksOutput, error)
	DescribeTasks(*ecs.DescribeTasksInput) (*ecs.DescribeTasksOutput, error)
}

// Scheduler is an implementation of the twelvefactor.Scheduler interface that
// is backed by ECS.
type Scheduler struct {
	// Cluster is the name of the ECS cluster to operate within. The zero
	// value is the "default" cluster.
	Cluster string

	ecs ecsClient

	// stackBuilder is the StackBuilder that will be used to provision AWS
	// resources.
	stackBuilder StackBuilder
}

// NewScheduler builds a new Scheduler instance backed by an ECS client
// that's configured with the given config.
func NewScheduler(config *aws.Config) *Scheduler {
	return &Scheduler{
		ecs:          ecs.New(config),
		stackBuilder: raw.NewStackBuilder(config),
	}
}

// Run creates or updates the associated ECS services for the individual
// processes within the application and runs them.
func (s *Scheduler) Run(app twelvefactor.App, processes ...twelvefactor.Process) error {
	return s.stackBuilder.Build(app, processes...)
}

// Remove removes the app and it's associated AWS resources.
func (s *Scheduler) Remove(app string) error {
	return s.stackBuilder.Remove(app)
}

// ScaleProcess scales the associated ECS service for the given app and process
// name.
func (s *Scheduler) ScaleProcess(app, process string, desired int) error {
	services, err := s.stackBuilder.Services(app)
	if err != nil {
		return err
	}

	// If there's no matching ECS service for this process, return an error.
	if _, ok := services[process]; !ok {
		return &ProcessNotFoundError{Process: process}
	}

	_, err = s.ecs.UpdateService(&ecs.UpdateServiceInput{
		Cluster:      aws.String(s.Cluster),
		DesiredCount: aws.Int64(int64(desired)),
		Service:      aws.String(services[process]),
	})
	return err
}

// Tasks returns the RUNNING and PENDING ECS tasks for the ECS services.
func (s *Scheduler) Tasks(app string) ([]twelvefactor.Task, error) {
	services, err := s.stackBuilder.Services(app)
	if err != nil {
		return nil, err
	}

	var tasks []twelvefactor.Task
	for _, service := range services {
		serviceTasks, err := s.ServiceTasks(service)
		if err != nil {
			return tasks, err
		}
		tasks = append(tasks, serviceTasks...)
	}

	return tasks, nil
}

// ServiceTasks returns the Tasks running for the given ECS service.
func (s *Scheduler) ServiceTasks(service string) ([]twelvefactor.Task, error) {
	listResp, err := s.ecs.ListTasks(&ecs.ListTasksInput{
		Cluster:     aws.String(s.Cluster),
		ServiceName: aws.String(service),
	})
	if err != nil {
		return nil, err
	}

	// No tasks.
	if len(listResp.TaskArns) == 0 {
		return nil, nil
	}

	describeResp, err := s.ecs.DescribeTasks(&ecs.DescribeTasksInput{
		Cluster: aws.String(s.Cluster),
		Tasks:   listResp.TaskArns,
	})
	if err != nil {
		return nil, err
	}

	var tasks []twelvefactor.Task
	for _, task := range describeResp.Tasks {
		id, err := arn.ResourceID(*task.TaskArn)
		if err != nil {
			return nil, err
		}

		tasks = append(tasks, twelvefactor.Task{
			ID:    id,
			State: *task.LastStatus,
		})
	}

	return tasks, nil
}
