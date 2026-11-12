# mcp-contract-tester: seven-day MVP

Goal: Test MCP tool contracts and failure behavior.

Status: scoped local MVP implemented; acceptance is limited to offline required/type checks and local initialization probing. See docs/DEMO.md for exclusions.

- [x] Day 1: Required-argument validation and fixtures.
- [x] Day 2: Validate supported JSON Schema types.
- [x] Day 3: Add local MCP transport adapter.
- [x] Day 4: Exercise timeouts and protocol errors.
- [x] Day 5: Generate actionable contract reports.
- [x] Day 6: Run against a deterministic example server.
- [x] Day 7: Package CLI, document scope and demonstrate failures.

## Definition of done

- A documented local command runs the scoped end-to-end example.
- Meaningful positive and failure-path tests pass locally and in CI.
- README distinguishes implemented features, mocks, measured results, and limitations.
- Example inputs and outputs are reproducible.
- A short demo script and evidence-based LinkedIn draft are saved locally.
- Real model quality is not claimed without a live evaluation.

Seven days is a target. Reduce scope or record unfinished work explicitly; never mark a project complete merely because its week ended.

## 9 November target-date assessment

MVP incomplete. Days 5–7 carry over: structured actionable reports, deterministic example server, packaged demo and evidence-based LinkedIn draft. Keep this project active until those acceptance checks pass; revised target 12 November, subject to validation.

12 November checkpoint: Day 5 reports are implemented and locally tested. Days 6–7 remain unfinished; the revised target is not met. Next run should implement the deterministic server and reproducible demo before reassessing completion.

## 12 November final local acceptance

All six definition-of-done items are satisfied locally: documented demo, regression checks, explicit limits, reproducible reports, demo script and unpublished LinkedIn draft. The demo asserts seven real CLI scenarios. CI now repeats the demo; portfolio completion must be recorded only after that run passes. Original seven-day target was missed; scoped MVP delivered on 12 November. Tool-call/schema interoperability remains future work, not completed scope.
