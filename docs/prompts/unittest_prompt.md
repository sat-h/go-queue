Write idiomatic, high-quality, and consistent unit tests for the following Go code. Follow these requirements:

* Use the standard testing package and the testify assertion library.
* Name test functions clearly, following Go conventions.
* Use table-driven tests for multiple scenarios.
* Use the actual struct definitions from the codebase in your tests. Do not define or use custom test structs with similar fields.
* Mock only true external dependencies (e.g., Redis, network calls, processors).
* Cover both success and failure paths, including edge cases.
* Avoid global state and ensure tests are independent.
* Use t.Parallel() where safe.
* Add comments for complex logic.
* Include tests for concurrency aspects (e.g., parallel job processing, graceful shutdown, race conditions) where relevant.
* Do not include integration tests or benchmarks.

If anything is unclear or there are multiple valid approaches, ask for clarification before proceeding.