# 12factor

12factor is a Go library for running [12factor](http://12factor.net/) applications.

## Terminology

### App

An App is a collection of Processes that share a common Environment.

### Process

A process is an individual 

## Packages

* [scheduler]: Provides an interface and various implementations for running 12factor apps. Implementations include Docker, ECS, Kubernetes and Nomad.
* [procfile]: Provides methods for parsing the Procfile manifest format.
