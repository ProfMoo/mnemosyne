# Linting

* After making changes, always run `make lint` before reporting completion, including for documentation and configuration changes.
* If you make further changes after linting, run `make lint` again.
* `make lint` must include the `go-lint` and `sql-lint` targets to lint Go and SQL files. Running only a targeted linter does not replace the full lint run.
* Fix lint violations introduced by your changes. Do not disable rules or add suppressions to bypass failures.
* If existing violations or environment issues prevent lint from passing, report the failing command and blockers explicitly. Never claim lint passed when it did not.
