package db

//go:generate go get -tool -modfile=sqlc.go.mod github.com/sqlc-dev/sqlc/cmd/sqlc@latest

// NOTE: this doesn't work, need to open a bug
// //go:generate go mod tidy -modfile=sqlc.go.mod

//go:generate go tool -modfile=sqlc.go.mod sqlc vet
//go:generate go tool -modfile=sqlc.go.mod sqlc generate

// ---

//go:generate go get -tool -modfile=tbls.go.mod github.com/k1LoW/tbls@latest

// NOTE: this doesn't work, need to open a bug
// //go:generate go mod tidy -modfile=tbls.go.mod

//go:generate go run ./newemptydb
//go:generate go tool -modfile=tbls.go.mod tbls doc --rm-dist
