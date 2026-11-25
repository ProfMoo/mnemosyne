# Configuration

* Singular source of config must be strongly typed YAML; use environment variables only where platform requires them (rarely, confirm with me first).
* `camelCase` keys unless the rest of the configuration consistently uses another case (ex: `kebab-case`).
* Use the `-config` command-line flag to select embedded configuration through the existing loader. Pass the full embedded filepath (for example, `-config config/local.yaml`), not an environment name or shorthand. Do not construct filenames from the flag value.
* Resolve Config Service references through config loader; avoid new env-dereferencing machinery.
* For new components, create a new top-level config object: `APIConfig` vs. `WorkerConfig`. Reusing smaller config blocks is allowed.
* Prefer `required:true` and updating all config files vs. `optional` and including `nil` checks in the code.
* Supply only the most specific config block that's needed for structs and modules (ex: don't pass in top-level `APIConfig` for a rate limiter. Just pass in a lower `RateLimitConfig`).

## Environment Configuration

* One component build for ALL environments. Don't do a build per environment.
* Avoid deployment-environment detection inside application code (ex: NEVER have something like `IsDevelopment()`).
* Environment filenames local to component: `config/development.yaml`, not `config/{component}-development.yaml`.

## Configuration Validation

* If validation of configuration is valuable, use `go-playground/validator` ONLY.
  * If custom validation is needed, use custom validation functions and loop it into the `go-playground/validator` call.
  * Do NOT validate config in factories (i.e. `New()`) or during application runtime.
