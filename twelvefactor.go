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
	// The command to run when running this process.
	Command string

	// Additional environment variables to merge with the App's environment
	// when running this process.
	Env map[string]string
}
