Much of the code here (everything in `sqlcgen`) is generated using `go generate` with `sqlc`.

https://www.alexedwards.net/blog/how-to-manage-tool-dependencies-in-go-1.24-plus

I decided to use a separate `modfile` (`go.tool.mod`) to separate `sqlc` dependencies needed at codegen time from `enventory` dependencies needed at build time. I created this file with:

```bash
go mod init -modfile=sqlc.go.mod github.com/bbkane/enventory
```

I've locked `sqlc` to the current latest version. To update it:

```bash
go get -tool -modfile=sqlc.go.mod github.com/sqlc-dev/sqlc/cmd/sqlc
```

And of course to re-generate code:

```bash
go generate ./...
```
