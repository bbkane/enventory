package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"go.bbkane.com/enventory/app/sqliteconnect/sqlcgen"
	"go.bbkane.com/enventory/models"
)

func (e *EnvService) varFindByID(ctx context.Context, id int64) (*models.Var, error) {
	queries := sqlcgen.New(e.dbtx)

	sqlcVar, err := queries.VarFindByID(ctx, id)
	if err != nil {
		return nil, models.ErrVarNotFound
	}

	return &models.Var{
		EnvName:    sqlcVar.EnvName,
		Name:       sqlcVar.Name,
		Comment:    sqlcVar.Comment,
		CreateTime: models.StringToTimeMust(sqlcVar.CreateTime),
		UpdateTime: models.StringToTimeMust(sqlcVar.UpdateTime),
		Value:      sqlcVar.Value,
		Enabled:    models.Int64ToBool(sqlcVar.Enabled),
	}, nil
}

func (e *EnvService) varFindID(ctx context.Context, envName string, name string) (int64, error) {
	queries := sqlcgen.New(e.dbtx)

	envID, err := e.envFindID(ctx, envName)
	if err != nil {
		return 0, err
	}

	id, err := queries.VarFindID(ctx, sqlcgen.VarFindIDParams{
		EnvID: envID,
		Name:  name,
	})

	if err != nil {
		return 0, models.ErrVarNotFound
	}
	return id, nil

}

func (e *EnvService) VarCreate(ctx context.Context, args models.VarCreateArgs) (*models.Var, error) {
	queries := sqlcgen.New(e.dbtx)

	envID, err := e.envFindID(ctx, args.EnvName)
	if err != nil {
		return nil, err
	}

	err = queries.VarCreate(ctx, sqlcgen.VarCreateParams{
		EnvID:      envID,
		Name:       args.Name,
		Comment:    args.Comment,
		CreateTime: models.TimeToString(args.CreateTime),
		UpdateTime: models.TimeToString(args.UpdateTime),
		Value:      args.Value,
		Enabled:    models.BoolToInt64(args.Enabled),
	})

	if err != nil {
		return nil, fmt.Errorf("could not create env var: %w", err)
	}
	return &models.Var{
		EnvName:    args.EnvName,
		Name:       args.Name,
		Comment:    args.Comment,
		CreateTime: args.CreateTime,
		UpdateTime: args.UpdateTime,
		Value:      args.Value,
		Enabled:    args.Enabled,
	}, nil
}

func (e *EnvService) VarDelete(ctx context.Context, envName string, name string) error {
	queries := sqlcgen.New(e.dbtx)

	envID, err := e.envFindID(ctx, envName)
	if err != nil {
		return err
	}

	rowsAffected, err := queries.VarDelete(ctx, sqlcgen.VarDeleteParams{
		EnvID: envID,
		Name:  name,
	})
	if err != nil {
		return fmt.Errorf("could not delete env var: %s: %s: %w", envName, name, err)
	}
	if rowsAffected == 0 {
		return models.ErrVarNotFound
	}
	return nil
}

func (e *EnvService) VarList(ctx context.Context, envName string) ([]models.Var, error) {
	queries := sqlcgen.New(e.dbtx)

	envID, err := e.envFindID(ctx, envName)
	if err != nil {
		return nil, err
	}

	envs, err := queries.VarList(ctx, envID)
	if err != nil {
		return nil, fmt.Errorf("could not list env vars: %s: %w", envName, err)
	}
	var ret []models.Var
	for _, sqlcEnv := range envs {
		ret = append(ret, models.Var{
			Name:       sqlcEnv.Name,
			Comment:    sqlcEnv.Comment,
			CreateTime: models.StringToTimeMust(sqlcEnv.CreateTime),
			EnvName:    envName,
			UpdateTime: models.StringToTimeMust(sqlcEnv.UpdateTime),
			Value:      sqlcEnv.Value,
			Enabled:    models.Int64ToBool(sqlcEnv.Enabled),
		})
	}

	return ret, nil
}

func (e *EnvService) VarShow(ctx context.Context, envName string, name string) (*models.Var, []models.VarRef, error) {
	queries := sqlcgen.New(e.dbtx)

	envID, err := e.envFindID(ctx, envName)
	if err != nil {
		return nil, nil, err
	}

	sqlEnvLocalVar, err := queries.VarShow(ctx, sqlcgen.VarShowParams{
		EnvID: envID,
		Name:  name,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("could not find env var: %s: %s: %w", envName, name, err)
	}

	envRefs := []models.VarRef{}
	sqlcEnvRefs, err := queries.VarRefListByVarID(ctx, sqlEnvLocalVar.VarID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, nil, err
	}

	for _, e := range sqlcEnvRefs {
		envRefs = append(envRefs, models.VarRef{
			EnvName:    e.EnvName,
			Name:       e.Name,
			Comment:    e.Comment,
			CreateTime: models.StringToTimeMust(e.CreateTime),
			UpdateTime: models.StringToTimeMust(e.UpdateTime),
			RefEnvName: envName,
			RevVarName: name,
			Enabled:    models.Int64ToBool(e.Enabled),
		})
	}

	return &models.Var{
		EnvName:    envName,
		Name:       name,
		Comment:    sqlEnvLocalVar.Comment,
		CreateTime: models.StringToTimeMust(sqlEnvLocalVar.CreateTime),
		UpdateTime: models.StringToTimeMust(sqlEnvLocalVar.UpdateTime),
		Value:      sqlEnvLocalVar.Value,
		Enabled:    models.Int64ToBool(sqlEnvLocalVar.Enabled),
	}, envRefs, nil
}

func (e *EnvService) VarUpdate(ctx context.Context, envName string, name string, args models.VarUpdateArgs) error {
	envVarID, err := e.varFindID(ctx, envName, name)
	if err != nil {
		return err
	}

	var newEnvID *int64
	if args.EnvName != nil {
		tmp, err := e.envFindID(ctx, *args.EnvName)
		if err != nil {
			return err
		}
		newEnvID = &tmp
	}

	queries := sqlcgen.New(e.dbtx)

	rowsAffected, err := queries.VarUpdate(ctx, sqlcgen.VarUpdateParams{
		EnvID:      newEnvID,
		Name:       args.Name,
		Comment:    args.Comment,
		CreateTime: models.TimePtrToStringPtr(args.CreateTime),
		UpdateTime: models.TimePtrToStringPtr(args.UpdateTime),
		Value:      args.Value,
		Enabled:    models.BoolPtrToInt64Ptr(args.Enabled),
		VarID:      envVarID,
	})

	if err != nil {
		return fmt.Errorf("err updating env var: %w", err)
	}
	if rowsAffected == 0 {
		return models.ErrVarNotFound
	}
	return nil
}
