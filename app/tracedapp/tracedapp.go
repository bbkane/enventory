package tracedapp

import (
	"context"
	"fmt"

	"go.bbkane.com/enventory/models"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

//nolint:gochecknoglobals  // this is a global tracer for the package
var Tracer = otel.Tracer("go.bbkane.com/enventory/tracedapp")

func ptrToString[T any](v *T) string {
	if v == nil {
		return "<nil>"
	}
	return fmt.Sprint(v)
}

type TracedApp struct {
	tracer trace.Tracer
	models.EnvService
}

func New(tracer trace.Tracer, envService models.EnvService) *TracedApp {
	return &TracedApp{
		tracer:     tracer,
		EnvService: envService,
	}
}

// -- Env

func (t *TracedApp) EnvCreate(ctx context.Context, args models.EnvCreateArgs) (*models.Env, error) {
	ctx, span := t.tracer.Start(
		ctx,
		"EnvCreate",
		trace.WithAttributes(
			attribute.String("args.Name", args.Name),
			attribute.String("args.Comment", args.Comment),
			attribute.String("args.CreateTime", models.TimeToString(args.CreateTime)),
			attribute.String("args.UpdateTime", models.TimeToString(args.UpdateTime)),
		),
	)
	defer span.End()

	env, err := t.EnvService.EnvCreate(ctx, args)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	return env, err
}

func (t *TracedApp) EnvDelete(ctx context.Context, name string) error {
	ctx, span := t.tracer.Start(ctx, "EnvDelete", trace.WithAttributes(attribute.String("name", name)))
	defer span.End()

	err := t.EnvService.EnvDelete(ctx, name)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	return err
}

func (t *TracedApp) EnvList(ctx context.Context, args models.EnvListArgs) ([]models.Env, error) {
	ctx, span := t.tracer.Start(ctx, "EnvList", trace.WithAttributes(
		attribute.String("args.Expr", ptrToString(args.Expr)),
	))
	defer span.End()

	envs, err := t.EnvService.EnvList(ctx, args)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	return envs, err
}

func (t *TracedApp) EnvUpdate(ctx context.Context, name string, args models.EnvUpdateArgs) error {
	ctx, span := t.tracer.Start(
		ctx,
		"EnvUpdate",
		trace.WithAttributes(
			attribute.String("name", name),
			attribute.String("args.Name", ptrToString(args.Name)),
			attribute.String("args.Comment", ptrToString(args.Comment)),
			attribute.String("args.CreateTime", ptrToString(models.TimePtrToStringPtr(args.CreateTime))),
			attribute.String("args.UpdateTime", ptrToString(models.TimePtrToStringPtr(args.UpdateTime))),
		),
	)
	defer span.End()

	err := t.EnvService.EnvUpdate(ctx, name, args)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	return err
}

func (t *TracedApp) EnvShow(ctx context.Context, name string) (*models.Env, error) {
	ctx, span := t.tracer.Start(ctx, "EnvShow", trace.WithAttributes(attribute.String("name", name)))
	defer span.End()

	env, err := t.EnvService.EnvShow(ctx, name)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	return env, err
}

// -- Var

func (t *TracedApp) VarCreate(ctx context.Context, args models.VarCreateArgs) (*models.Var, error) {
	ctx, span := t.tracer.Start(
		ctx,
		"VarCreate",
		trace.WithAttributes(
			attribute.String("args.EnvName", args.EnvName),
			attribute.String("args.Name", args.Name),
			attribute.String("args.Value", args.Value),
			attribute.String("args.Comment", args.Comment),
			attribute.String("args.CreateTime", models.TimeToString(args.CreateTime)),
			attribute.String("args.UpdateTime", models.TimeToString(args.UpdateTime)),
		),
	)
	defer span.End()

	variable, err := t.EnvService.VarCreate(ctx, args)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	return variable, err
}

func (t *TracedApp) VarDelete(ctx context.Context, envName string, name string) error {
	ctx, span := t.tracer.Start(
		ctx,
		"VarDelete",
		trace.WithAttributes(
			attribute.String("envName", envName),
			attribute.String("name", name),
		),
	)
	defer span.End()

	err := t.EnvService.VarDelete(ctx, envName, name)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	return err
}

func (t *TracedApp) VarList(ctx context.Context, envName string) ([]models.Var, error) {
	ctx, span := t.tracer.Start(
		ctx,
		"VarList",
		trace.WithAttributes(attribute.String("envName", envName)),
	)
	defer span.End()

	vars, err := t.EnvService.VarList(ctx, envName)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	return vars, err
}

func (t *TracedApp) VarUpdate(ctx context.Context, envName string, name string, args models.VarUpdateArgs) error {
	ctx, span := t.tracer.Start(
		ctx,
		"VarUpdate",
		trace.WithAttributes(
			attribute.String("envName", envName),
			attribute.String("name", name),
			attribute.String("args.Name", ptrToString(args.Name)),
			attribute.String("args.Value", ptrToString(args.Value)),
			attribute.String("args.Comment", ptrToString(args.Comment)),
			attribute.String("args.CreateTime", ptrToString(models.TimePtrToStringPtr(args.CreateTime))),
			attribute.String("args.UpdateTime", ptrToString(models.TimePtrToStringPtr(args.UpdateTime))),
		),
	)
	defer span.End()

	err := t.EnvService.VarUpdate(ctx, envName, name, args)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	return err
}

func (t *TracedApp) VarShow(ctx context.Context, envName string, name string) (*models.Var, []models.VarRef, error) {
	ctx, span := t.tracer.Start(
		ctx,
		"VarShow",
		trace.WithAttributes(
			attribute.String("envName", envName),
			attribute.String("name", name),
		),
	)
	defer span.End()

	variable, refs, err := t.EnvService.VarShow(ctx, envName, name)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	return variable, refs, err
}

// -- VarRef

func (t *TracedApp) VarRefCreate(ctx context.Context, args models.VarRefCreateArgs) (*models.VarRef, error) {
	ctx, span := t.tracer.Start(
		ctx,
		"VarRefCreate",
		trace.WithAttributes(
			attribute.String("args.EnvName", args.EnvName),
			attribute.String("args.Name", args.Name),
			attribute.String("args.Comment", args.Comment),
			attribute.String("args.CreateTime", models.TimeToString(args.CreateTime)),
			attribute.String("args.UpdateTime", models.TimeToString(args.UpdateTime)),
			attribute.String("args.RefEnvName", args.RefEnvName),
			attribute.String("args.RefVarName", args.RefVarName),
		),
	)
	defer span.End()

	varRef, err := t.EnvService.VarRefCreate(ctx, args)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	return varRef, err
}

func (t *TracedApp) VarRefDelete(ctx context.Context, envName string, name string) error {
	ctx, span := t.tracer.Start(
		ctx,
		"VarRefDelete",
		trace.WithAttributes(
			attribute.String("envName", envName),
			attribute.String("name", name),
		),
	)
	defer span.End()

	err := t.EnvService.VarRefDelete(ctx, envName, name)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	return err
}

func (t *TracedApp) VarRefList(ctx context.Context, envName string) ([]models.VarRef, []models.Var, error) {
	ctx, span := t.tracer.Start(
		ctx,
		"VarRefList",
		trace.WithAttributes(attribute.String("envName", envName)),
	)
	defer span.End()

	varRefs, vars, err := t.EnvService.VarRefList(ctx, envName)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	return varRefs, vars, err
}

func (t *TracedApp) VarRefShow(ctx context.Context, envName string, name string) (*models.VarRef, *models.Var, error) {
	ctx, span := t.tracer.Start(
		ctx,
		"VarRefShow",
		trace.WithAttributes(
			attribute.String("envName", envName),
			attribute.String("name", name),
		),
	)
	defer span.End()

	varRef, varRefVar, err := t.EnvService.VarRefShow(ctx, envName, name)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	return varRef, varRefVar, err
}

func (t *TracedApp) VarRefUpdate(ctx context.Context, envName string, name string, args models.VarRefUpdateArgs) error {
	ctx, span := t.tracer.Start(
		ctx,
		"VarRefUpdate",
		trace.WithAttributes(
			attribute.String("envName", envName),
			attribute.String("name", name),
			attribute.String("args.Name", ptrToString(args.Name)),
			attribute.String("args.Comment", ptrToString(args.Comment)),
			attribute.String("args.CreateTime", ptrToString(models.TimePtrToStringPtr(args.CreateTime))),
			attribute.String("args.UpdateTime", ptrToString(models.TimePtrToStringPtr(args.UpdateTime))),
			attribute.String("args.RefEnvName", ptrToString(args.RefEnvName)),
			attribute.String("args.RefVarName", ptrToString(args.RefVarName)),
		),
	)
	defer span.End()

	err := t.EnvService.VarRefUpdate(ctx, envName, name, args)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	return err
}

// -- WithTx

func (t *TracedApp) WithTx(ctx context.Context, fn func(es models.EnvService) error) error {
	ctx, span := t.tracer.Start(ctx, "WithTx")
	defer span.End()

	err := t.EnvService.WithTx(ctx, func(es models.EnvService) error {
		// Wrap the EnvService with tracing.
		tracedES := New(t.tracer, es)

		return fn(tracedES)
	})

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return err
}
