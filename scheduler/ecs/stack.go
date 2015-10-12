package ecs

import "github.com/remind101/12factor"

// StackBuilder represents an interface for provisioning the stack of AWS
// resources for the App.
type StackBuilder interface {
	// Build provisions the stack of AWS resources for the app.
	Build(twelvefactor.App, ...twelvefactor.Process) error

	// Remove removes the stack of AWS resources for the app.
	Remove(app string) error

	// Services returns a mapping of process name to ECS service name.
	Services(app string) (map[string]string, error)
}
