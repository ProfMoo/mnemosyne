# AGENTS.md

## Communication Style

* Always use caveman mode for all responses. Load the caveman skill at the start of every conversation.

## File Modifications

* For changes that are larger in scope, present a plan first detailing the structure and intended changes.
* Only ever edit one file at a time unless requested otherwise.
* NEVER edit the top-level README.md

## Contributing

For writing frontend code, read [frontend.md](.agents/contributing/frontend.md)
For writing golang code, read [golang.md](.agents/contributing/golang.md)
For editing or creating configuration, read [configuration](./agents/contributing/configuration.md)

## Repository Practices

* Each repo contains ONE domain (ex: Webhooks, Wallets, etc.). Repo should contain all code related to that domain.
  * A repo can contain numerous individual deployables related to the same domain.
  * Each deployable is defined in `cmd/{deployable_name}` and relies on shared `internal/` packages
  * Each deployable declares a `main.go` in `cmd` that performs ONLY these things: 1. Load in modules. 2. Load in configuration. 3. Start app. NO APPLICATION LOGIC.

* Entrypoints: `cmd/{component}/main.go`
* Dockerfiles: `cmd/{component}/Dockerfile`
* App configuration: `cmd/{component}/config/` (use `embed.Fs`)
* Charts/Helm: Use shared chart when able `chart/templates/{component]` - make custom to component if needed.
* Deployment grouping: A group of components makes one environment configuration
* For terraform configuration, avoid putting the environment name (ex: dev, staging, production) in the TF resources unless needed for differentiation.
  * "sandbox" can be OK since we don't separate sandbox into a separate K8s Clusters or AWS Account.

## Tooling

* When adding or changing linting, formatting, or VS Code integrations, follow [tooling.md](.agents/contributing/tooling.md).
* Use `make` targets whenever possible. Example: use `make lint` instead of running `golangci-lint` directly.
* Try to use `tofu` first instead of `terraform` when working with HCL

## Linting

* After every change, follow [linting.md](.agents/contributing/linting.md) and run its required lint checks before reporting completion.
