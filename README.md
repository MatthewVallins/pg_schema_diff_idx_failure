# pg-schema-diff Expression Index Failure

Demonstrates [stripe/pg-schema-diff](https://github.com/stripe/pg-schema-diff) failing on expression-based indexes.

## Prerequisites

- Go 1.21+
- Docker

## Run

```bash
./run_demo.sh
```

## Issue

Without `WithNoConcurrentIndexOps()`, the library cannot even generate the initial index creation statement for expression-based indexes.

With `WithNoConcurrentIndexOps()`, index generation works but the internal validation step still fails because the option does not propagate. The only workaround is to disable validation entirely with `WithDoNotValidatePlan()`, which defeats the purpose of safe migration generation.
