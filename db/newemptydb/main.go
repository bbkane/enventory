package main

import (
	"context"

	"go.bbkane.com/enventory/db"
)

func main() {
	ctx := context.Background()
	_, err := db.Connect(ctx, "tmp.db")
	if err != nil {
		panic(err)
	}
}
