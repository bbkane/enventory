---
mode: agent
tools: ['codebase']
---

Add a `var ref update` command. It should look like:

```bash
enventory var ref update \
    --comment 'newcomment' \
    --confirm true \
    --create-time <time> \
    --db-path <path> \
    --env <env> \
    --name myrefid \
    --new-env <another env> \
    --new-name myrefidnewname \
    --ref-env <env> \
    --ref-var <ref var name> \
    --timeout <timeout> \
    --update-time <time>
```

This should be very similar to `var update` or `env update` commands

- Add a `VarRefUpdate` method to `EnvService` in `models/env.go` that calls the `sqlcgen.VarRefUpdate` method
- Implement it in `app/var_ref.go`
- create `VarRefUpdateCmd` and `varRefUpdateRun` in `cli/var_ref.go`
- Add a test in `main_var_ref_test.go`