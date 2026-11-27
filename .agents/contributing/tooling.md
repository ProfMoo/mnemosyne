# Linting, Formatting, and Editor Tooling

## Tool Selection and Configuration

* When adding a language or changing its developer tooling, provide both linting and formatting support. Reuse existing tools before introducing new ones.
* Pin tool versions using the repository's existing dependency management patterns. Use `mise.toml` for standalone tools and Go tool dependencies for Go-native tools where appropriate.
* Keep linting and formatting configuration in version-controlled repository files. CLI tools and editor integrations must use the same configuration and managed tool version.
* Use strict rules appropriate to the language and dialect. Do not weaken rules or suppress failures to make existing files pass.

## Make Targets

* Follow the existing `<language>-lint` and `<language>-fmt` target naming pattern, such as `sql-lint` and `sql-fmt`.
* Add language-specific targets as prerequisites of the aggregate `lint` and `fmt` targets. Mark all these targets `.PHONY`.
* Keep linting read-only and make violations return a nonzero exit code. Formatting targets may rewrite files.
* If formatting checks exist, extend them to cover the new formatter without modifying files.
* Use managed tool execution, such as `mise exec -- sqlfluff`, rather than depending on an unpinned global installation.

## VS Code

* Configure the language-specific `editor.defaultFormatter` and enable `editor.formatOnSave` in `.vscode/settings.json`.
* Point the extension at the repository configuration and managed executable. Prefer portable paths or supported variables; do not introduce user-specific absolute paths.
* Add required extensions to `.vscode/extensions.json` recommendations. Verify extension identifiers and supported settings rather than guessing them.
* Preserve unrelated editor settings. Do not disable another language's tooling when configuring a new one.
* Check whether required extensions are installed. Recommendations alone do not install them; report any installation needed to activate the integration.

## Verification

* Verify aggregate Make targets invoke the intended tools. Exercise formatting on sample input or explicitly scoped files without rewriting unrelated user work.
* Validate editor configuration and executable resolution. Distinguish CLI verification from an actual editor format-on-save test.
* Follow [linting.md](linting.md): run `make lint` after changes and report any failures or unverified editor behavior explicitly.
