**Before writing any tests, always check for ambiguities or multiple valid approaches. If anything is unclear, ask for clarification before proceeding.**

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
* **Include tests for concurrency aspects:**
    - Ensure multiple jobs can be processed in parallel, verifying correct concurrent behavior and throughput.
    - Test graceful shutdown, confirming no jobs are left unprocessed, no panics occur, and resources are cleaned up.
    - Check for race conditions or timing-related bugs, and ensure tests can be run with the `-race` flag.
* Do not include integration tests or benchmarks.
* Do not redefine or duplicate any types (structs, interfaces, etc.) from the codebase in your test files. Always use the actual types by importing them.
