package twelvefactor

import "io"

// Runner is an interface that wraps the basic Run method, providing a way to
// run a 12factor application.
type Runner interface {
	Run(*App) error
}

// DetachedRunner is an optional interface for running a detached processes. This
// is useful for running long running management tasks, such as database
// migrations.
type DetachedRunner interface {
	RunDetached(app string, process Process) error
}

// AttachedRunner is an optional interface for running attached processes. This
// is usefuly for running attached processes like a rails console.
type AttachedRunner interface {
	RunAttached(app string, process Process, in io.Reader, out io.Writer) error
}

// Scaler is an interface that wraps the basic Scale method for scaling a
// process by name for an application.
type Scaler interface {
	Scale(app, process string, desired int) error
}

// Remover is an interface that wraps the basic Remove method for removing an
// app and all of it's processes.
type Remover interface {
	Remove(app) error
}

// Scheduler provides an interface for running twelve factor applications.
type Scheduler interface {
	Runner
	Scaler
	Remover
}
