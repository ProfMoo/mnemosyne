# Golang

## Dependency Injection

* Uses **Uber Fx** for dependency injection. Each main.go file defines fx.App with modules from `internal/` packages.
* Never use Fx result tags; define distinct interfaces or types instead.
* Don't make useless 'provide' methods to facilitate Fx. Instead, use the canonical singular `NewFoo()` method or use `fx.As`, `fx.Decorate`, or other Fx functionality.
* Use `fx.In` and a dedicated `NewFooParams` struct.
* Always leave a newline between `fx.In`/`fx.Out` and the rest of the param. Example:

  ```go
  type myParams struct {
    fx.In

    OtherThings OtherThings
  }
  ```

## Testing Requirements

* Use **testify** test suites with suite struct per file
* Tests in `{package}_test` package (separate from code under test)
* Use suite assert methods for assertions
* Follow workflow: write code → write tests → verify `make test` passes
* Only test public methods
* Never make a method public solely to test it; test behavior through existing public interfaces
* `go test ./...` reports `[no test files]` for tagged packages. Verify with `make unit`, or `go test -tags=unit ./...` — a bare `go test` run proves nothing

## Build Tags

* Identify if the repo supports marking test scope by build tags. To do this, look at `make` commands for tag refs and scan existing tests for this block:

  ```go
  //go:build unit
  // +build unit
  ```

  If build tags are used broadly, then ensure each unit test has the correct unit test tag, such as:

  ```go
  //go:build unit
  // +build unit
  ```

  and that each integration test has the correct tag, such as:

  ```go
  //go:build integration
  // +build integration
  ```

  Defer to the Makefile and existing conventions to find the correct tags.

* If you need real dependencies for writing integration tests, use `testcontainers`. Do not use `docker-compose`. Example: MongoDB, Kafka, Clickhouse.

## Test value bar

Coverage is not the goal; a test that fails when behavior breaks is. Before adding a test, confirm it would catch a real regression:

* Assert on observable behavior, not on internal call sequences or struct internals. A test that must change whenever the implementation is refactored has negative value
* Don't restructure existing passing tests without a clear benefit — the diff costs reviewers their existing understanding of them
* Prefer tests on discrete, re-used packages over glue code. Ex: prefer testing core business logic, not database calls/clients.
* Never assert on `time.Now()` reads; Instead, insert a stubbed `time.Now()` and modify the return value to test desired behavior.

## Mock Generation

* **Use mockery.yaml** (not mockgen) - add any new interfaces that need mocking to top-level `.mockery.yaml`
* **Never** write mocks manually
* Generate with `make` commands - don't use mockery directly
* Generated mocks are **testify mocks**. Use the `testify/mock` API against them: `mock.Anything`, `mock.MatchedBy`, `.On(...).Return(...)`, `.AssertExpectations(t)`

## Interfaces

* Don't add a method onto an existing interface unless it 100% makes sense to do so (ex: a user of the interface will always use both methods). Most cases, make a new interface instead.
* Interface should be nouns, methods should be verbs. Interface names are short and to the point. Example:

  ```go
  type ThingDoer interface {
    DoThing(input1, input2) (result, error)
  }
  ```

## Observability

* ALL observability (ex: metrics, logging, traces) should be completed in an interceptor.
* Each interceptor should serve JUST ONE observability purpose.
* The request/response structure of the method call should provide enough information to observe appropriately.

```go
type ThingDoer interface {
 HandleMatches(ctx context.Context, matches []*events.Match) (outcomes []*events.Match, err error)
}

type ThingDoerFunc func(context.Context, []*events.Match) ([]*events.Match, error)

func (f ThingDoerFunc) HandleMatches(ctx context.Context, matches []*events.Match) ([]*events.Match, error) {
 return f(ctx, matches)
}

type ThingDoerMetrics struct {
 stats stats.Client
}

func NewThingDoerMetrics(stats stats.Client) *ThingDoerMetrics {
 return &ThingDoerMetrics{stats: stats}
}

func (m *ThingDoerMetrics) HandleMatches(next ThingDoer) ThingDoerFunc {
 return func(ctx context.Context, matches []*events.Match) ([]*events.Match, error) {
  ...handle metrics here
 }
}

type ThingDoerTracer struct {
 tracer tracing.TracerProvider
}

func NewThingDoerTracer(tracer tracing.TracerProvider) *ThingDoerTracer {
 return &ThingDoerTracer{tracer: tracer}
}

func (t *ThingDoerTracer) HandleMatches(next ThingDoer) ThingDoerFunc {
 return func(ctx context.Context, matches []*events.Match) ([]*events.Match, error) {
  ...handle tracing here
 }
}

type NewThingDoerInterceptorChainParams struct {
 fx.In

 Base         *ThingDoer
 Metrics      *ThingDoerMetrics
 Tracer       *ThingDoerTracer
}

func NewThingDoerInterceptorChain(params NewThingDoerInterceptorChainParams) ThingDoer {
 return params.Tracer.HandleMatches(params.Metrics.HandleMatches(params.Base))
}
```

## Architecture

## Data

These rules must be followed when fetching or creating data to/from any data layer

* The domain-layer must own the types and repository interface. Example:

  ```go
  type Thing struct {
    ThingyPropertyOne string
    ThingyPropertyTwo string
  }

  type ThingRepository interface {
    GetThing(ctx context.context, thingId string) (Thing, error)
  }
  ```

  Then, the data layer implements that interface and return the data-layer defined type. The interface must be generically implementable by different data storage technologies.

## API

These rules must be followed when fetching or creating data to/from the API surface

* The domain-layer must own the type and conversion to the API types
* The domain-layer must own any default values handling, page token handling, and anything that will maintain if the database or API expositional format changes
* API Handler should be *business-logic only*. Any generic logic (billing, rate limiting, etc.) should be done in middleware.
  * Use middleware before the handler for all generic HTTP observability: metrics, logs, traces.

## Code Quality

* Do not use mutexes unless definitely needed
* **No comments** unless absolutely necessary. Examples of where it's necessary: the behavior of the code is surprising, the decisions made during development need explanation.
  * If you include a comment, it MUST CONTAIN information that's relevant an on ongoing basis. Don't include information that's only relevant in the context of a PR (ex: migration information).
* Keep files under 200-300 lines
* When fixing a bug, exhaust options in the existing implementation before introducing a new pattern or technology
* Avoid writing one-off scripts as committed files
* If you are replacing code, **remove the old code** alongside the new code
* *Never* pass important values through the `ctx`.

## Source Code Organization

* Organize files by "importance", descending. Start with consts, errors. Then the important interfaces and structs. Then the methods on the structs.
  * The file should "read" from top-to-bottom - when a method is mentioned in another method, that method should be just below it
  * When making a receiver method, put the receiver method below the struct definition
* Never create generic `interfaces.go`, `types.go`, `consts.go` files; use a name or location describing their domain purpose
* Each package should be a discrete domain, such as: `address`, `urlvalidator`, or `errors`, `httpserver`.
* All new packages in `internal` should be at the top-level. Do not create nested packages (i.e. `internal/foo/bar/baz`).
* Make functions, structs, and interface private by default. Only make them public if they are used by another package.
* Place private (i.e. lowercase) consts in the scope closest to where it's used. If it's just used in one method, make it a const in that method.  If it's used just in one file, make it a const in that file.
* Place publicly available errors in the file that it's used, near the top of the file. It's part of the public interface.

## Miscellaneous

* NEVER pass domain context through `context.Context`.
* NEVER use an Abstract Factory pattern. Use composition and the Dispatcher pattern instead.
