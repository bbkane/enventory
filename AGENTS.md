# Instructions for AGENTs for specific tasks

## Architecture Overview

- Top level "presentation" layer - cli package
- Business layer in the middle - app package
- Generated DB layer with sqlc - sqlcgen package
- cli and app packages share types via a models package

## Making a change

Follow the current steps in order:

Back up current DB

```bash
cp ~/.config/enventory.db  ~/.config/enventory.db.$(date +'%Y-%m-%d.%H.%M.%S').bak"
```

If you need to to make a SQL change:

Add a SQL migrations if necessary in ./db/migrations

Update SQL query if necessary in ./db/queries/

Generate Go code to call the SQL query: `go generate ./...`

Update models.Service interface with a new arg or new method

Update models.TracedService implementation to emit trace information for the new thing

Update app.Service implementation to call the SQL Go code inside a transaction (`WithTx`).

Update ./cli/ to the new thing (new flag, new command, etc.)

Update output functons

Add a snapshot test (see other tests for examples) and run it