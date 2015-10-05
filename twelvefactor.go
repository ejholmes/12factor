// Package twelvefactor provides types to represents 12factor applications,
// which are defined in http://12factor.net/
package twelvefactor

// App represents a 12factor application. We define an application has a
// collection of processes that share a common environment.
type App struct {
	// Unique identifier of the application.
	ID string

	// Name of the application.
	Name string

	// The container image for this app.
	Image string

	// The shared environment variables for the individual processes.
	Env map[string]string

	// The Processes that compose this application.
	Processes []Process
}

// Process represents an individual Process of an App, which defines the image
// to run as well as the command.
type Process struct {
	// A unique identifier for this process, within the scope of the app.
	// Generally this would be something like "web" or "worker.
	Name string

	// The command to run when running this process.
	Command string

	// Additional environment variables to merge with the App's environment
	// when running this process.
	Env map[string]string

	// Free form labels to attach to this process.
	Labels map[string]string

	// Where Stdout for this process should go to.
	Stdout Stdout

	// Where Stdin for this process should come from. The zero value is to
	// not attach Stdin.
	Stdin Stdin

	// The desired number of instances to run.
	DesiredCount int

	// The amount of memory to allocate to this process, in bytes.
	MemoryLimit int

	// The number of CPU Shares to allocate to this process.
	CPUShares int
}

// Stdout is an interface that represents a the location to send Stdout to.
type Stdout interface{}

// Stdin represents the location to get Stdin from.
type Stdin interface{}
