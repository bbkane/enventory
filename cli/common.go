package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"go.bbkane.com/enventory/app"
	"go.bbkane.com/enventory/models"
	"go.bbkane.com/motel"
	"go.bbkane.com/warg"
	"go.bbkane.com/warg/completion"
	"go.bbkane.com/warg/path"
	"go.bbkane.com/warg/value/contained"
	"go.bbkane.com/warg/value/scalar"
	"go.opentelemetry.io/otel"
	"golang.org/x/term"
)

//nolint:gochecknoglobals // cwd will not change
var cwd string

//nolint:gochecknoglobals // global tracer
var tracer = otel.Tracer("go.bbkane.com/enventory/cli")

func init() { //nolint:gochecknoinits  // cwd will not change
	var err error
	cwd, err = os.Getwd()
	if err != nil {
		// I don't know when this could happen?
		panic(err)
	}
}

func emptyOrNil[T any](iFace interface{}) (T, error) {
	under, ok := iFace.(T)
	if !ok {
		return under, contained.ErrIncompatibleInterface
	}
	return under, nil
}

// datetime is a type for the CLI so I can pass strings in and parse them to dates
func datetime() contained.TypeInfo[time.Time] {
	return contained.TypeInfo[time.Time]{
		Description: "datetime in RFC3339 format",
		FromIFace:   emptyOrNil[time.Time],
		FromString: func(s string) (time.Time, error) {
			return time.Parse(time.RFC3339, s)
		},
		FromZero: func() time.Time {
			return time.Time{}
		},
	}
}

func confirmFlag() warg.FlagMap {
	return warg.FlagMap{
		"--confirm": warg.NewFlag(
			"Ask for confirmation before running",
			scalar.Bool(
				scalar.Default(true),
			),
			warg.Required(),
		),
	}
}

func maskFlag() warg.FlagMap {
	return warg.FlagMap{
		"--mask": warg.NewFlag(
			"Mask values when printing",
			scalar.Bool(
				scalar.Default(true),
			),
			warg.EnvVars("ENVELOPE_MASK"),
			warg.Required(),
		),
	}
}

func formatFlag() warg.FlagMap {
	return warg.FlagMap{
		"--format": warg.NewFlag(
			"output format",
			scalar.String(
				scalar.Choices("table", "value-only"),
				scalar.Default("table"),
			),
			warg.Required(),
		),
	}
}

func widthFlag() warg.FlagMap {

	// TODO: figure out a good way to cache this for all width flags
	width := 0
	if term.IsTerminal(0) {
		termWidth, _, err := term.GetSize(0)
		if err == nil { // if there's not an error
			width = termWidth
		}
	}

	return warg.FlagMap{
		"--width": warg.NewFlag(
			"Width of the table. 0 means no limit",
			scalar.Int(
				scalar.Default(width),
			),
			warg.Required(),
		),
	}
}

func completeExistingEnvName(ctx context.Context, es models.Service, cmdCtx warg.CmdContext) (*completion.Candidates, error) {
	// TODO: should this use expr?
	envs, err := es.EnvList(ctx, models.EnvListArgs{Expr: nil})
	if err != nil {
		return nil, fmt.Errorf("could not list envs for completion: %w", err)
	}
	candidates := &completion.Candidates{
		Type:   completion.Type_ValuesDescriptions,
		Values: nil,
	}
	for _, e := range envs {
		candidates.Values = append(candidates.Values, completion.Candidate{
			Name:        e.Name,
			Description: e.Comment,
		})
	}
	return candidates, nil
}

func envNameFlag() warg.Flag {
	return warg.NewFlag(
		"Environment name",
		scalar.String(
			scalar.Default(cwd),
		),
		warg.Required(),
		warg.FlagCompletions(withEnvServiceCompletions(
			completeExistingEnvName)),
	)
}

func completeExistingEnvVarName(
	ctx context.Context, es models.Service, cmdCtx warg.CmdContext) (*completion.Candidates, error) {
	// no completions if we can't get the env name
	envNamePtr := ptrFromMap[string](cmdCtx.Flags, "--env")
	if envNamePtr == nil {
		return nil, nil
	}

	vars, err := es.VarList(ctx, *envNamePtr)
	if err != nil {
		return nil, fmt.Errorf("could not get env for completion: %w", err)
	}
	candidates := &completion.Candidates{
		Type:   completion.Type_ValuesDescriptions,
		Values: nil,
	}
	for _, v := range vars {
		candidates.Values = append(candidates.Values, completion.Candidate{
			Name:        v.Name,
			Description: v.Comment,
		})
	}
	return candidates, nil
}

func completeExistingRefEnvVarName(
	ctx context.Context, es models.Service, cmdCtx warg.CmdContext) (*completion.Candidates, error) {
	// this is quite copy paste (just changed the flag name from --env to --ref-env), but I think I'm feeling lazy
	envNamePtr := ptrFromMap[string](cmdCtx.Flags, "--ref-env")
	if envNamePtr == nil {
		return nil, nil
	}

	vars, err := es.VarList(ctx, *envNamePtr)
	if err != nil {
		return nil, fmt.Errorf("could not get env for completion: %w", err)
	}
	candidates := &completion.Candidates{
		Type:   completion.Type_ValuesDescriptions,
		Values: nil,
	}
	for _, v := range vars {
		candidates.Values = append(candidates.Values, completion.Candidate{
			Name:        v.Name,
			Description: v.Comment,
		})
	}
	return candidates, nil
}

func varNameFlag() warg.Flag {
	return warg.NewFlag(
		"Env var name",
		scalar.String(),
		warg.Required(),
		warg.FlagCompletions(withEnvServiceCompletions(
			completeExistingEnvVarName)),
	)
}

func completeExistingVarRefName(
	ctx context.Context, es models.Service, cmdCtx warg.CmdContext) (*completion.Candidates, error) {
	// no completions if we can't get the env name
	envNamePtr := ptrFromMap[string](cmdCtx.Flags, "--env")
	if envNamePtr == nil {
		return nil, nil
	}

	varRefs, _, err := es.VarRefList(ctx, *envNamePtr)
	if err != nil {
		return nil, fmt.Errorf("could not get env for completion: %w", err)
	}
	candidates := &completion.Candidates{
		Type:   completion.Type_ValuesDescriptions,
		Values: nil,
	}
	for _, v := range varRefs {
		candidates.Values = append(candidates.Values, completion.Candidate{
			Name:        v.Name,
			Description: v.Comment,
		})
	}
	return candidates, nil
}

func varRefNameFlag() warg.Flag {
	return warg.NewFlag(
		"Var ref name",
		scalar.String(),
		warg.Required(),
		warg.FlagCompletions(withEnvServiceCompletions(
			completeExistingVarRefName)),
	)
}

func sqliteDSNFlagMap() warg.FlagMap {

	return warg.FlagMap{
		"--db-path": warg.NewFlag(
			"Sqlite DSN. Usually the file name",
			scalar.Path(
				scalar.Default(path.New("~/.config/enventory.db")),
			),
			warg.Required(),
			warg.EnvVars("ENVELOPE_DB_PATH"),
		),
	}
}

func commonCreateFlagMapPtrs(comment *string, createTime *time.Time, updateTime *time.Time) warg.FlagMap {
	now := time.Now()
	commonCreateFlags := warg.FlagMap{
		"--comment": warg.NewFlag(
			"Comment",
			scalar.String(
				scalar.Default(""),
				scalar.PointerTo(comment),
			),
			warg.Required(),
		),
		"--create-time": warg.NewFlag(
			"Create time",
			scalar.New(
				datetime(),
				scalar.Default(now),
				scalar.PointerTo(createTime),
			),
			warg.Required(),
		),
		"--update-time": warg.NewFlag(
			"Update time",
			scalar.New(
				datetime(),
				scalar.Default(now),
				scalar.PointerTo(updateTime),
			),
			warg.Required(),
		),
	}
	return commonCreateFlags
}

func commonCreateFlagMap() warg.FlagMap {
	now := time.Now()
	commonCreateFlags := warg.FlagMap{
		"--comment": warg.NewFlag(
			"Comment",
			scalar.String(
				scalar.Default(""),
			),
			warg.Required(),
		),
		"--create-time": warg.NewFlag(
			"Create time",
			scalar.New(
				datetime(),
				scalar.Default(now),
			),
			warg.Required(),
		),
		"--update-time": warg.NewFlag(
			"Update time",
			scalar.New(
				datetime(),
				scalar.Default(now),
			),
			warg.Required(),
		),
	}
	return commonCreateFlags
}

func commonUpdateFlags() warg.FlagMap {

	commonUpdateFlags := warg.FlagMap{
		"--comment": warg.NewFlag(
			"Comment",
			scalar.String(),
		),
		"--create-time": warg.NewFlag(
			"Create time",
			scalar.New(
				datetime(),
			),
		),
		"--new-name": warg.NewFlag(
			"New name",
			scalar.String(),
		),
		"--update-time": warg.NewFlag(
			"Update time",
			scalar.New(
				datetime(),
				scalar.Default(time.Now()),
			),
			warg.UnsetSentinel("UNSET"),
		),
	}
	return commonUpdateFlags
}

func timeoutFlagMap() warg.FlagMap {
	timeoutFlag := warg.FlagMap{
		"--timeout": warg.NewFlag(
			"Timeout for a run. Use https://pkg.go.dev/time#Duration to build it",
			scalar.Duration(
				scalar.Default(10*time.Minute),
			),
			warg.Required(),
		),
	}
	return timeoutFlag

}

func timeZoneFlagMap() warg.FlagMap {
	return warg.FlagMap{
		"--timezone": warg.NewFlag(
			"Timezone to display dates",
			scalar.String(
				scalar.Default("local"),
				scalar.Choices("local", "utc"),
			),
			warg.Required(),
		),
	}
}

// ptrFromMap returns &val if key is in the map, otherwise nil
// useful for converting from the cmdCtx.Flags to the types domain needs
func ptrFromMap[T any](m map[string]any, key string) *T {
	val, exists := m[key]
	if exists {
		ret := val.(T)
		return &ret
	}
	return nil
}

type commonCreateArgs struct {
	Comment    string
	CreateTime time.Time
	UpdateTime time.Time
}

func mustGetCommonCreateArgs(pf warg.PassedFlags) commonCreateArgs {
	return commonCreateArgs{
		Comment:    pf["--comment"].(string),
		CreateTime: pf["--create-time"].(time.Time),
		UpdateTime: pf["--update-time"].(time.Time),
	}
}

type commonUpdateArgs struct {
	Comment    *string
	CreateTime *time.Time
	NewName    *string
	UpdateTime *time.Time
}

func getCommonUpdateArgs(pf warg.PassedFlags) commonUpdateArgs {
	return commonUpdateArgs{
		Comment:    ptrFromMap[string](pf, "--comment"),
		CreateTime: ptrFromMap[time.Time](pf, "--create-time"),
		NewName:    ptrFromMap[string](pf, "--new-name"),
		UpdateTime: ptrFromMap[time.Time](pf, "--update-time"),
	}
}

func mustGetEnvNameArg(pf warg.PassedFlags) string {
	return pf["--env"].(string)
}

func mustGetMaskArg(pf warg.PassedFlags) bool {
	return pf["--mask"].(bool)
}

func mustGetNameArg(pf warg.PassedFlags) string {
	return pf["--name"].(string)
}

func mustGetTimeoutArg(pf warg.PassedFlags) time.Duration {
	return pf["--timeout"].(time.Duration)
}

func mustGetTimezoneArg(pf warg.PassedFlags) string {
	return pf["--timezone"].(string)
}

func mustGetWidthArg(pf warg.PassedFlags) int {
	return pf["--width"].(int)
}

// withSetup wraps a cli.Action to read --db-path and --timeout and creates
//   - a context from the timeout
//   - a tracer provider (and sets it globally)
//   - an EnvService ()
func withSetup(
	f func(ctx context.Context, es models.Service, cmdCtx warg.CmdContext) error,
) warg.Action {
	return func(cmdCtx warg.CmdContext) error {

		ctx, cancel := context.WithTimeout(
			context.Background(),
			mustGetTimeoutArg(cmdCtx.Flags),
		)
		defer cancel()

		tracerProvider, err := motel.NewTracerProviderFromEnv(ctx, motel.NewTracerProviderFromEnvArgs{
			AppName: cmdCtx.App.Name,
			Version: cmdCtx.App.Version,
		})

		if err != nil {
			return fmt.Errorf("could not init tracerProvider: %w", err)
		}
		otel.SetTracerProvider(tracerProvider)
		defer func() {
			// best effort reporting if shutdown fails
			err = tracerProvider.Shutdown(ctx)
			if err != nil {
				fmt.Fprintf(os.Stderr, "could not shutdown tracer: %v", err)
			}
		}()

		nameParts := make([]string, len(cmdCtx.ParseState.SectionPath)+1)
		copy(nameParts, cmdCtx.ParseState.SectionPath)
		nameParts[len(nameParts)-1] = cmdCtx.ParseState.CurrentCmdName

		rootSpanName := strings.Join(nameParts, " ")
		ctx, rootSpan := tracer.Start(ctx, rootSpanName)
		defer rootSpan.End()

		ctx, span := tracer.Start(ctx, "withSetup")
		defer span.End()

		sqliteDSN := cmdCtx.Flags["--db-path"].(path.Path).MustExpand()
		es, err := app.NewEnvService(ctx, sqliteDSN)
		es = models.NewTracedService(models.Tracer, es)
		if err != nil {
			return fmt.Errorf("could not create env service: %w", err)
		}

		return f(ctx, es, cmdCtx)
	}
}

// withEnvService wraps a cli.Action to read --db-path and --timeout and create a EnvService
func withEnvServiceCompletions(
	f func(ctx context.Context, es models.Service, cmdCtx warg.CmdContext) (*completion.Candidates, error),
) warg.CompletionsFunc {
	return func(cmdCtx warg.CmdContext) (*completion.Candidates, error) {

		ctx, cancel := context.WithTimeout(
			context.Background(),
			mustGetTimeoutArg(cmdCtx.Flags),
		)
		defer cancel()

		sqliteDSN := cmdCtx.Flags["--db-path"].(path.Path).MustExpand()
		es, err := app.NewEnvService(ctx, sqliteDSN)
		if err != nil {
			return nil, fmt.Errorf("could not create env service: %w", err)
		}

		return f(ctx, es, cmdCtx)
	}
}

// withConfirm wraps a cli.Action to ask for confirmation before running
func withConfirm(f func(cmdCtx warg.CmdContext) error) warg.Action {
	return func(cmdCtx warg.CmdContext) error {
		confirm := cmdCtx.Flags["--confirm"].(bool)
		if !confirm {
			return f(cmdCtx)
		}

		fmt.Print("Type 'yes' to continue: ")
		reader := bufio.NewReader(os.Stdin)
		confirmation, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("confirmation ReadString error: %w", err)
		}
		confirmation = strings.TrimSpace(confirmation)
		if confirmation != "yes" {
			return fmt.Errorf("unconfirmed change")
		}
		return f(cmdCtx)
	}
}
