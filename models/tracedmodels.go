package models

import (
	"context"
	"database/sql"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

//nolint:gochecknoglobals  // this is a global tracer for the package
var Tracer = otel.Tracer("go.bbkane.com/enventory/models")

func ptrToString[T any](v *T) string {
	// TODO: this could probably be faster...
	if v == nil {
		return "<nil>"
	}
	return fmt.Sprint(v)
}

// -- TracedDBTX
type TracedDBTX struct {
	tracer trace.Tracer
	DBTX
}

func NewTracedDBTX(tracer trace.Tracer, dbtx DBTX) *TracedDBTX {
	return &TracedDBTX{
		tracer: tracer,
		DBTX:   dbtx,
	}
}

var _ DBTX = (*TracedDBTX)(nil)

func (t *TracedDBTX) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	ctx, span := t.tracer.Start(ctx, "ExecContext", trace.WithAttributes(
		attribute.String("query", query),
	))
	defer span.End()
	result, err := t.DBTX.ExecContext(ctx, query, args...)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return result, err
}

func (t *TracedDBTX) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	ctx, span := t.tracer.Start(ctx, "PrepareContext", trace.WithAttributes(
		attribute.String("query", query),
	))
	defer span.End()
	stmt, err := t.DBTX.PrepareContext(ctx, query)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return stmt, err
}

func (t *TracedDBTX) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	ctx, span := t.tracer.Start(ctx, "QueryContext", trace.WithAttributes(
		attribute.String("query", query),
	))
	defer span.End()
	rows, err := t.DBTX.QueryContext(ctx, query, args...)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return rows, err
}

func (t *TracedDBTX) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	ctx, span := t.tracer.Start(ctx, "QueryRowContext", trace.WithAttributes(
		attribute.String("query", query),
	))
	defer span.End()
	row := t.DBTX.QueryRowContext(ctx, query, args...)
	if row.Err() != nil {
		span.RecordError(row.Err())
		span.SetStatus(codes.Error, row.Err().Error())
	}
	return row
}

// Unwrap the TracedDBTX to get the underlying DBTX. Needed to start a transaction.
func (t *TracedDBTX) Unwrap() DBTX {
	return t.DBTX
}

// -- TracedService

type TracedService struct {
	tracer trace.Tracer
	Service
}

var _ Service = (*TracedService)(nil)

func NewTracedService(tracer trace.Tracer, envService Service) *TracedService {
	return &TracedService{
		tracer:  tracer,
		Service: envService,
	}
}

// -- Env

func (t *TracedService) EnvCreate(ctx context.Context, args EnvCreateArgs) (*Env, error) {
	ctx, span := t.tracer.Start(
		ctx,
		"EnvCreate",
		trace.WithAttributes(
			attribute.String("args.Name", args.Name),
			attribute.String("args.Comment", args.Comment),
			attribute.String("args.CreateTime", TimeToString(args.CreateTime)),
			attribute.String("args.UpdateTime", TimeToString(args.UpdateTime)),
			attribute.Bool("args.Enabled", args.Enabled),
		),
	)
	defer span.End()

	env, err := t.Service.EnvCreate(ctx, args)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	return env, err
}

func (t *TracedService) EnvDelete(ctx context.Context, name string) error {
	ctx, span := t.tracer.Start(ctx, "EnvDelete", trace.WithAttributes(attribute.String("name", name)))
	defer span.End()

	err := t.Service.EnvDelete(ctx, name)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	return err
}

func (t *TracedService) EnvList(ctx context.Context, args EnvListArgs) ([]Env, error) {
	ctx, span := t.tracer.Start(ctx, "EnvList", trace.WithAttributes(
		attribute.String("args.Expr", ptrToString(args.Expr)),
	))
	defer span.End()

	envs, err := t.Service.EnvList(ctx, args)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	return envs, err
}

func (t *TracedService) EnvUpdate(ctx context.Context, name string, args EnvUpdateArgs) error {
	ctx, span := t.tracer.Start(
		ctx,
		"EnvUpdate",
		trace.WithAttributes(
			attribute.String("name", name),
			attribute.String("args.Name", ptrToString(args.Name)),
			attribute.String("args.Comment", ptrToString(args.Comment)),
			attribute.String("args.CreateTime", ptrToString(TimePtrToStringPtr(args.CreateTime))),
			attribute.String("args.UpdateTime", ptrToString(TimePtrToStringPtr(args.UpdateTime))),
			attribute.String("args.Enabled", ptrToString(args.Enabled)),
		),
	)
	defer span.End()

	err := t.Service.EnvUpdate(ctx, name, args)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	return err
}

func (t *TracedService) EnvShow(ctx context.Context, name string) (*Env, error) {
	ctx, span := t.tracer.Start(ctx, "EnvShow", trace.WithAttributes(attribute.String("name", name)))
	defer span.End()

	env, err := t.Service.EnvShow(ctx, name)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	return env, err
}

// -- Var

func (t *TracedService) VarCreate(ctx context.Context, args VarCreateArgs) (*Var, error) {
	ctx, span := t.tracer.Start(
		ctx,
		"VarCreate",
		trace.WithAttributes(
			attribute.String("args.EnvName", args.EnvName),
			attribute.String("args.Name", args.Name),
			attribute.String("args.Value", "<redacted>"), // can be sensitive
			attribute.String("args.Comment", args.Comment),
			attribute.String("args.CreateTime", TimeToString(args.CreateTime)),
			attribute.String("args.UpdateTime", TimeToString(args.UpdateTime)),
			attribute.Bool("args.Enabled", args.Enabled),
		),
	)
	defer span.End()

	variable, err := t.Service.VarCreate(ctx, args)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	return variable, err
}

func (t *TracedService) VarDelete(ctx context.Context, envName string, name string) error {
	ctx, span := t.tracer.Start(
		ctx,
		"VarDelete",
		trace.WithAttributes(
			attribute.String("envName", envName),
			attribute.String("name", name),
		),
	)
	defer span.End()

	err := t.Service.VarDelete(ctx, envName, name)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	return err
}

func (t *TracedService) VarList(ctx context.Context, envName string) ([]Var, error) {
	ctx, span := t.tracer.Start(
		ctx,
		"VarList",
		trace.WithAttributes(attribute.String("envName", envName)),
	)
	defer span.End()

	vars, err := t.Service.VarList(ctx, envName)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	return vars, err
}

func (t *TracedService) VarUpdate(ctx context.Context, envName string, name string, args VarUpdateArgs) error {
	argsValue := "<nil>"
	if args.Value != nil {
		argsValue = "<redacted>" // can be sensitive
	}
	ctx, span := t.tracer.Start(
		ctx,
		"VarUpdate",
		trace.WithAttributes(
			attribute.String("envName", envName),
			attribute.String("name", name),
			attribute.String("args.Name", ptrToString(args.Name)),
			attribute.String("args.Value", argsValue), // can be sensitive
			attribute.String("args.Comment", ptrToString(args.Comment)),
			attribute.String("args.CreateTime", ptrToString(TimePtrToStringPtr(args.CreateTime))),
			attribute.String("args.UpdateTime", ptrToString(TimePtrToStringPtr(args.UpdateTime))),
			attribute.String("args.Enabled", ptrToString(args.Enabled)),
		),
	)
	defer span.End()

	err := t.Service.VarUpdate(ctx, envName, name, args)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	return err
}

func (t *TracedService) VarShow(ctx context.Context, envName string, name string) (*Var, []VarRef, error) {
	ctx, span := t.tracer.Start(
		ctx,
		"VarShow",
		trace.WithAttributes(
			attribute.String("envName", envName),
			attribute.String("name", name),
		),
	)
	defer span.End()

	variable, refs, err := t.Service.VarShow(ctx, envName, name)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	return variable, refs, err
}

// -- VarRef

func (t *TracedService) VarRefCreate(ctx context.Context, args VarRefCreateArgs) (*VarRef, error) {
	ctx, span := t.tracer.Start(
		ctx,
		"VarRefCreate",
		trace.WithAttributes(
			attribute.String("args.EnvName", args.EnvName),
			attribute.String("args.Name", args.Name),
			attribute.String("args.Comment", args.Comment),
			attribute.String("args.CreateTime", TimeToString(args.CreateTime)),
			attribute.String("args.UpdateTime", TimeToString(args.UpdateTime)),
			attribute.String("args.RefEnvName", args.RefEnvName),
			attribute.String("args.RefVarName", args.RefVarName),
			attribute.Bool("args.Enabled", args.Enabled),
		),
	)
	defer span.End()

	varRef, err := t.Service.VarRefCreate(ctx, args)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	return varRef, err
}

func (t *TracedService) VarRefDelete(ctx context.Context, envName string, name string) error {
	ctx, span := t.tracer.Start(
		ctx,
		"VarRefDelete",
		trace.WithAttributes(
			attribute.String("envName", envName),
			attribute.String("name", name),
		),
	)
	defer span.End()

	err := t.Service.VarRefDelete(ctx, envName, name)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	return err
}

func (t *TracedService) VarRefList(ctx context.Context, envName string) ([]VarRef, []Var, error) {
	ctx, span := t.tracer.Start(
		ctx,
		"VarRefList",
		trace.WithAttributes(attribute.String("envName", envName)),
	)
	defer span.End()

	varRefs, vars, err := t.Service.VarRefList(ctx, envName)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	return varRefs, vars, err
}

func (t *TracedService) VarRefShow(ctx context.Context, envName string, name string) (*VarRef, *Var, error) {
	ctx, span := t.tracer.Start(
		ctx,
		"VarRefShow",
		trace.WithAttributes(
			attribute.String("envName", envName),
			attribute.String("name", name),
		),
	)
	defer span.End()

	varRef, varRefVar, err := t.Service.VarRefShow(ctx, envName, name)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	return varRef, varRefVar, err
}

func (t *TracedService) VarRefUpdate(ctx context.Context, envName string, name string, args VarRefUpdateArgs) error {
	ctx, span := t.tracer.Start(
		ctx,
		"VarRefUpdate",
		trace.WithAttributes(
			attribute.String("envName", envName),
			attribute.String("name", name),
			attribute.String("args.Name", ptrToString(args.Name)),
			attribute.String("args.Comment", ptrToString(args.Comment)),
			attribute.String("args.CreateTime", ptrToString(TimePtrToStringPtr(args.CreateTime))),
			attribute.String("args.UpdateTime", ptrToString(TimePtrToStringPtr(args.UpdateTime))),
			attribute.String("args.RefEnvName", ptrToString(args.RefEnvName)),
			attribute.String("args.RefVarName", ptrToString(args.RefVarName)),
			attribute.String("args.Enabled", ptrToString(args.Enabled)),
		),
	)
	defer span.End()

	err := t.Service.VarRefUpdate(ctx, envName, name, args)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	return err
}

// -- WithTx

func (t *TracedService) WithTx(ctx context.Context, fn func(ctx context.Context, es Service) error) error {
	ctx, span := t.tracer.Start(ctx, "WithTx")
	defer span.End()

	err := t.Service.WithTx(ctx, func(ctx context.Context, es Service) error {
		// Wrap the EnvService with tracing.
		tracedES := NewTracedService(t.tracer, es)

		return fn(ctx, tracedES)
	})

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return err
}
