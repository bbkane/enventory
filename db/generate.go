package db

//go:generate go get -tool -modfile=go.sqlc.mod github.com/sqlc-dev/sqlc/cmd/sqlc@latest

// NOTE: this doesn't work, need to open a bug
// //go:generate go mod tidy -modfile=go.sqlc.mod

//go:generate go tool -modfile=go.sqlc.mod sqlc vet
//go:generate go tool -modfile=go.sqlc.mod sqlc generate
