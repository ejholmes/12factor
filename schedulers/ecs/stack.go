package ecs

import "github.com/remind101/12factor"

// StackBuilder represents an interface for provisioning the stack of AWS
// resources for the App.
type StackBuilder interface {
	// Build provisions the stack of AWS resources for the app.
	Build(twelvefactor.App) error

	// Remove removes the stack of AWS resources for the app.
	Remove(app string) error

	// Service returns the name of the ECS services for the given process
	// name.
	Service(app, process string) (string, error)
}
