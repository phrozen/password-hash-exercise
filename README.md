[![Go Reference](https://pkg.go.dev/badge/github.com/phrozen/password-hash-exercise.svg)](https://pkg.go.dev/github.com/phrozen/password-hash-exercise)
[![Go Report Card](https://goreportcard.com/badge/github.com/phrozen/password-hash-exercise)](https://goreportcard.com/report/github.com/phrozen/password-hash-exercise)

# password-hash-exercise

The purpose of the exercise is to provide a service that hashes and saves passwords with SHA512. It is a small project but implemented loosely with hexagonal/clean architecture as much as possible, as to not to over enginner the solution by providing too much abstraction.

Project structure follows best practices: https://github.com/golang-standards/project-layout

There are 4 main layers:

### Store

Store implements an interface to `Set` or `Get` raw byte data, the `id` to retrieve the data is provided by the store when saving (implementation dependant) and there are no search, list, update or delete operations provided by design (requirements).

A single in `memory` implementation is provided, that has configurable delayed writes and makes use of Go `atomic` and `sync` packages to handle concurrency, while storage is backed by a `map`. Other implementations can be provided later and easily swapped.

A useful `Close` method is required, as most data stores require some sort of teardown process to ensure data integrity (like pending writes, ongoing connections, etc...), some implementations might not need it, but it is such a common scenario, that those implementations can mock it.

Package: `/internal/memory`

### Application

This is the core business logic, an `application` is responsible for the actual use case of taking a password, SHA512 hashing, base64 encoding, and saving and retrieving from the provided `Store`.

There is not much to is, because the actual requirement is very simple, but it perfectly encapsulates what is needed as the core feature of the project.

Package: `/internal/app`

### Service

Provides the REST API layer for the application, it manages all the routing, handlers and errors for the application, as well as shutdown signaling for the `Server`.

It also provides a simple logging middleware for testign and debugging purposes that can be easily modified to fit other needs. This `Logger` implements a response observer pattern to keep track of response codes sent by the `Service` and can easily be built upon to track response sizes and other meaningful data for the logs.

`Stats` are handled by the service, as the assumption is that it is not a core business requirement, it was made as an *ad-hoc* feature for the purposes of the exercise, but in real life scenarios, should be either moved to the application as core business logic, or as an additional middleware that tracks everything inside the service.

Packages:

+ `/internal/service`
+ `/internal/stats`
+ `/internal/middleware/logger`

### Server (http)

Handles all HTTP protocol related communications like transport and listening for connections, and pass them down to the service. It is also responsible for handling graceful shutdown as the `main` executable for the application. The `Server` interface is provided along with the main project as implementation was so small it didn't warrant its own package.

It receives a single `Service` to handle all request routing and processing, and it listens to the service shutdown signal as well as operating system shutdown signals seamlessly.

Package: `/cmd/server`

## Testing

Most of the packages have 100% coverage, and all `tests` and relevant `benchmarks` are done using only the standard library. Packages where 100% coverage cannot be achieved are clearly documented on branches where unreachable code and error cases that cannot fail are found. The code still is left as is (defensive coding) for robustness and to avoid a change in downstream implementation breaking the application.

## Final thoughts

Although Go standard library is extensive and well polished, it suffers from the `1.x` backwards compatibility problem, as well as the idea from the core team that if better packages exist, that functionality does not need to be retrofitted into the standard library. Due to this, you can find some areas of opportunity when creating a Go project like this, where well established third party libraries are clearly the best call (and best practice).

In my opinion, as a minimum, a modern/performant router should be used instead of `http.ServeMux`, mainly because of performance and code readability, there are many popular ones around and any would be fine. Another place where a third party library would make code more maintainable is in testing, there is a reason `testify` is widely considered an extension to the standard library, the `testing` package is great, but testify's assert and mock just enhances it for the best.

Lastly, for further optimizations, the use of a framework for the service layer and an additional abstraction will benefit code maintainability in the long run, and can be easily decoupled if implemented correctly, as well as configuration libraries and real world data stores (even embeded ones like `bolt`).

All in all, it was a good exercise, but some stuff became tedious fast due to the amount of boilerplate. Go projects should be aimed at what the language does best, code readability and maintainability, most stack decisions should be made towards that goal for projects to prosper.