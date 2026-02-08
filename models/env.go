package models

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

// -- Env

var ErrEnvNotFound = errors.New("env not found")

type Env struct {
	Name       string
	Comment    string
	CreateTime time.Time
	UpdateTime time.Time
	Enabled    bool
}

type EnvCreateArgs struct {
	Name       string
	Comment    string
	CreateTime time.Time
	UpdateTime time.Time
	Enabled    bool
}

type EnvListArgs struct {
	Expr *string
}

type EnvUpdateArgs struct {
	Comment    *string
	CreateTime *time.Time
	Name       *string
	UpdateTime *time.Time
	Enabled    *bool
}

// -- Var

var ErrVarNotFound = errors.New("local var not found")

type Var struct {
	EnvName     string
	Name        string
	Comment     string
	CreateTime  time.Time
	UpdateTime  time.Time
	Value       string
	Enabled     bool
	Completions []string
}

type VarCreateArgs struct {
	EnvName     string
	Name        string
	Comment     string
	CreateTime  time.Time
	UpdateTime  time.Time
	Value       string
	Enabled     bool
	Completions []string
}

type VarUpdateArgs struct {
	Comment     *string
	CreateTime  *time.Time
	EnvName     *string
	Name        *string
	UpdateTime  *time.Time
	Value       *string
	Enabled     *bool
	Completions *[]string
}

// -- VarRef

var ErrVarRefNotFound = errors.New("local ref not found")

type VarRef struct {
	EnvName    string
	Name       string
	Comment    string
	CreateTime time.Time
	UpdateTime time.Time
	RefEnvName string
	RevVarName string
	Enabled    bool
}

type VarRefCreateArgs struct {
	EnvName    string
	Name       string
	Comment    string
	CreateTime time.Time
	UpdateTime time.Time
	RefEnvName string
	RefVarName string
	Enabled    bool
}

type VarRefUpdateArgs struct {
	Comment    *string
	CreateTime *time.Time
	EnvName    *string // for --new-env
	Name       *string // for --new-name
	UpdateTime *time.Time
	RefEnvName *string // for --ref-env
	RefVarName *string // for --ref-var
	Enabled    *bool
}

// -- EnvExportable

type EnvExportable struct {
	Name    string
	Enabled bool
	Value   string
}

// -- interface

type DBTX interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}

type Service interface {
	EnvCreate(ctx context.Context, args EnvCreateArgs) (*Env, error)
	EnvDelete(ctx context.Context, name string) error
	EnvList(ctx context.Context, args EnvListArgs) ([]Env, error)
	EnvUpdate(ctx context.Context, name string, args EnvUpdateArgs) error
	EnvShow(ctx context.Context, name string) (*Env, error)

	EnvExportableList(ctx context.Context, envName string) ([]EnvExportable, error)

	// TODO: should envName be its own parameter?
	VarCreate(ctx context.Context, args VarCreateArgs) (*Var, error)
	VarDelete(ctx context.Context, envName string, name string) error
	VarList(ctx context.Context, envName string) ([]Var, error)
	VarUpdate(ctx context.Context, envName string, name string, args VarUpdateArgs) error
	VarShow(ctx context.Context, envName string, name string) (*Var, []VarRef, error)

	VarRefCreate(ctx context.Context, args VarRefCreateArgs) (*VarRef, error)
	VarRefDelete(ctx context.Context, envName string, name string) error
	VarRefList(ctx context.Context, envName string) ([]VarRef, []Var, error)
	VarRefShow(ctx context.Context, envName string, name string) (*VarRef, *Var, error)
	VarRefUpdate(ctx context.Context, envName string, name string, args VarRefUpdateArgs) error

	WithTx(ctx context.Context, fn func(ctx context.Context, es Service) error) error
}
