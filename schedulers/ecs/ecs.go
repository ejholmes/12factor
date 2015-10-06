// Package ecs provides a scheduler for running 12factor applications using ECS.
package ecs

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/remind101/12factor"
)

// ecsClient represents a client for interacting with ECS.
type ecsClient interface {
	UpdateService(*ecs.UpdateServiceInput) (*ecs.UpdateServiceOutput, error)
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

// NewScheduler returns a new Scheduler instance backed by the given ECS client.
func NewScheduler(c *ecs.ECS) *Scheduler {
	return &Scheduler{
		ecs: c,
	}
}

// Run creates or updates the associated ECS services for the individual
// processes within the application and runs them.
func (s *Scheduler) Run(app twelvefactor.App) error {
	return s.stackBuilder.Build(app)
}

// Remove removes the app and it's associated AWS resources.
func (s *Scheduler) Remove(app string) error {
	return s.stackBuilder.Remove(app)
}

// ScaleProcess scales the associated ECS service for the given app and process
// name.
func (s *Scheduler) ScaleProcess(app, process string, desired int) error {
	service, err := s.stackBuilder.Service(app, process)
	if err != nil {
		return err
	}
	_, err = s.ecs.UpdateService(&ecs.UpdateServiceInput{
		Cluster:      aws.String(s.Cluster),
		DesiredCount: aws.Int64(int64(desired)),
		Service:      aws.String(service),
	})
	return err
}
