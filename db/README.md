Much of the code here (everything in `sqlcgen` and `dbdoc`) is generated using `go generate` with `sqlc` and `tbls`

https://www.alexedwards.net/blog/how-to-manage-tool-dependencies-in-go-1.24-plus

I decided to use a separate `modfile` (`go.<tool>.mod`) to separate dependencies for tools needed at codegen time from `enventory` dependencies needed at build time. I created this file with:


```bash
# sqlc
go mod init -modfile=go.sqlc.mod github.com/bbkane/sqlc-tool

# tbls
go mod init -modfile=go.tbls.mod github.com/bbkane/tbls-tool
```


And of course to re-generate code:

```bash
go generate ./...
```

