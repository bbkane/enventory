package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/expr-lang/expr"
	"github.com/xhit/go-str2duration/v2"
	"go.bbkane.com/enventory/app/sqliteconnect/sqlcgen"
	"go.bbkane.com/enventory/models"
)

// mapErrEnvNotFound replaces sql.ErrNoRows with domain.ErrEnvNotFound but otherwise
// passes it through.
//
// Deprecated: I want to replace this with envFindID, but that'll require rewriting some sql
func mapErrEnvNotFound(e error) error {
	if errors.Is(e, sql.ErrNoRows) {
		return models.ErrEnvNotFound
	} else {
		return e
	}
}

// envFindID looks for en env's SQLite ID and returns a wrapped ErrEnvNotFound or a sql error
func (e *EnvService) envFindID(ctx context.Context, envName string) (int64, error) {
	queries := sqlcgen.New(e.dbtx)
	envID, err := queries.EnvFindID(ctx, envName)
	if errors.Is(err, sql.ErrNoRows) {
		err = models.ErrEnvNotFound
	}
	if err != nil {
		return 0, fmt.Errorf("could not find env with name: %s: %w", envName, err)
	}
	return envID, nil
}

func (e *EnvService) EnvCreate(ctx context.Context, args models.EnvCreateArgs) (*models.Env, error) {
	queries := sqlcgen.New(e.dbtx)

	createdEnvRow, err := queries.EnvCreate(ctx, sqlcgen.EnvCreateParams{
		Name:       args.Name,
		Comment:    args.Comment,
		CreateTime: models.TimeToString(args.CreateTime),
		UpdateTime: models.TimeToString(args.UpdateTime),
	})

	if err != nil {
		return nil, fmt.Errorf("could not create env in db: %w", err)
	}

	return &models.Env{
		Name:       createdEnvRow.Name,
		Comment:    createdEnvRow.Comment,
		CreateTime: models.StringToTimeMust(createdEnvRow.CreateTime),
		UpdateTime: models.StringToTimeMust(createdEnvRow.UpdateTime),
	}, nil
}

func (e *EnvService) EnvDelete(ctx context.Context, name string) error {
	queries := sqlcgen.New(e.dbtx)

	rowsAffected, err := queries.EnvDelete(ctx, name)
	if err != nil {
		return mapErrEnvNotFound(err)
	}
	if rowsAffected == 0 {
		return models.ErrEnvNotFound
	}
	return nil
}

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (e *EnvService) EnvList(ctx context.Context, args models.EnvListArgs) ([]models.Env, error) {
	queries := sqlcgen.New(e.dbtx)

	sqlcEnvs, err := queries.EnvList(ctx)
	if err != nil {
		return nil, err
	}

	ret := []models.Env{}
	for _, e := range sqlcEnvs {
		ret = append(ret, models.Env{
			Name:       e.Name,
			Comment:    e.Comment,
			CreateTime: models.StringToTimeMust(e.CreateTime),
			UpdateTime: models.StringToTimeMust(e.UpdateTime),
		})
	}

	// TODO: use --timezone since I'm messing with dates
	// go run . env list --expr 'filter(Envs, pathExists(.Name) and .CreateTime > now() - duration("52w"))'
	pathExists := expr.Function(
		"pathExists",
		func(params ...any) (any, error) {
			return pathExists(params[0].(string))
		},
		new(func(string) (bool, error)),
	)

	duration := expr.Function(
		"duration",
		func(params ...any) (any, error) {
			return str2duration.ParseDuration(params[0].(string))
		},
		str2duration.ParseDuration,
	)

	if args.Expr != nil {
		program, err := expr.Compile(*args.Expr, duration, pathExists)
		if err != nil {
			return nil, fmt.Errorf("could not compile expr: %w", err)
		}

		exprEnv := map[string]any{
			"Envs":       ret,
			"pathExists": pathExists,
		}

		output, err := expr.Run(program, exprEnv)
		if err != nil {
			return nil, fmt.Errorf("could not eval expr: %w", err)
		}
		if filteredEnvs, ok := output.([]any); ok {
			ret = nil
			for _, e := range filteredEnvs {
				if env, ok := e.(models.Env); ok {
					ret = append(ret, env)
				} else {
					return nil, fmt.Errorf("expr slice item is not a models.Env: %T: %#v", e, e)
				}
			}
		} else {
			return nil, fmt.Errorf("expr output is not a slice: %T", output)
		}
	}

	return ret, nil
}

func (e *EnvService) EnvUpdate(ctx context.Context, name string, args models.EnvUpdateArgs) error {

	queries := sqlcgen.New(e.dbtx)

	rowsAffected, err := queries.EnvUpdate(ctx, sqlcgen.EnvUpdateParams{
		NewName:    args.Name,
		Comment:    args.Comment,
		CreateTime: models.TimePtrToStringPtr(args.CreateTime),
		UpdateTime: models.TimePtrToStringPtr(args.UpdateTime),
		Name:       name,
	})

	if err != nil {
		return fmt.Errorf("err updating env: %w", mapErrEnvNotFound(err))
	}
	if rowsAffected == 0 {
		return models.ErrEnvNotFound
	}

	return nil
}

func (e *EnvService) EnvShow(ctx context.Context, name string) (*models.Env, error) {
	queries := sqlcgen.New(e.dbtx)

	sqlcEnv, err := queries.EnvShow(ctx, name)

	if err != nil {
		return nil, mapErrEnvNotFound(err)
	}

	return &models.Env{
		Name:       name,
		Comment:    sqlcEnv.Comment,
		CreateTime: models.StringToTimeMust(sqlcEnv.CreateTime),
		UpdateTime: models.StringToTimeMust(sqlcEnv.UpdateTime),
	}, nil
}
