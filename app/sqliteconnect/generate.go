package sqliteconnect

//go:generate go get -tool -modfile=go.tool.mod github.com/sqlc-dev/sqlc/cmd/sqlc@latest
//go:generate go tool -modfile=go.tool.mod sqlc vet
//go:generate go tool -modfile=go.tool.mod sqlc generate
