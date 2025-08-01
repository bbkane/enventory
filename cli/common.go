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
	"go.bbkane.com/warg/completion"
	"go.bbkane.com/warg/flag"
	"go.bbkane.com/warg/path"
	"go.bbkane.com/warg/wargcore"
	"go.opentelemetry.io/otel"
	"golang.org/x/term"

	"go.bbkane.com/warg/value/contained"
	"go.bbkane.com/warg/value/scalar"
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
		Empty: func() time.Time {
			return time.Time{}
		},
	}
}

func confirmFlag() wargcore.FlagMap {
	return wargcore.FlagMap{
		"--confirm": flag.New(
			"Ask for confirmation before running",
			scalar.Bool(
				scalar.Default(true),
			),
			flag.Required(),
		),
	}
}

func maskFlag() wargcore.FlagMap {
	return wargcore.FlagMap{
		"--mask": flag.New(
			"Mask values when printing",
			scalar.Bool(
				scalar.Default(true),
			),
			flag.EnvVars("ENVELOPE_MASK"),
			flag.Required(),
		),
	}
}

func formatFlag() wargcore.FlagMap {
	return wargcore.FlagMap{
		"--format": flag.New(
			"output format",
			scalar.String(
				scalar.Choices("table", "value-only"),
				scalar.Default("table"),
			),
			flag.Required(),
		),
	}
}

func widthFlag() wargcore.FlagMap {

	// TODO: figure out a good way to cache this for all width flags
	width := 0
	if term.IsTerminal(0) {
		termWidth, _, err := term.GetSize(0)
		if err == nil { // if there's not an error
			width = termWidth
		}
	}

	return wargcore.FlagMap{
		"--width": flag.New(
			"Width of the table. 0 means no limit",
			scalar.Int(
				scalar.Default(width),
			),
			flag.Required(),
		),
	}
}

func completeExistingEnvName(ctx context.Context, es models.Service, cmdCtx wargcore.Context) (*completion.Candidates, error) {
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

func envNameFlag() wargcore.Flag {
	return flag.New(
		"Environment name",
		scalar.String(
			scalar.Default(cwd),
		),
		flag.Required(),
		flag.CompletionCandidates(withEnvServiceCompletions(
			completeExistingEnvName)),
	)
}

func completeExistingEnvVarName(
	ctx context.Context, es models.Service, cmdCtx wargcore.Context) (*completion.Candidates, error) {
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
	ctx context.Context, es models.Service, cmdCtx wargcore.Context) (*completion.Candidates, error) {
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

func varNameFlag() wargcore.Flag {
	return flag.New(
		"Env var name",
		scalar.String(),
		flag.Required(),
		flag.CompletionCandidates(withEnvServiceCompletions(
			completeExistingEnvVarName)),
	)
}

func completeExistingVarRefName(
	ctx context.Context, es models.Service, cmdCtx wargcore.Context) (*completion.Candidates, error) {
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

func varRefNameFlag() wargcore.Flag {
	return flag.New(
		"Var ref name",
		scalar.String(),
		flag.Required(),
		flag.CompletionCandidates(withEnvServiceCompletions(
			completeExistingVarRefName)),
	)
}

func sqliteDSNFlagMap() wargcore.FlagMap {

	return wargcore.FlagMap{
		"--db-path": flag.New(
			"Sqlite DSN. Usually the file name",
			scalar.Path(
				scalar.Default(path.New("~/.config/enventory.db")),
			),
			flag.Required(),
			flag.EnvVars("ENVELOPE_DB_PATH"),
		),
	}
}

func commonCreateFlagMapPtrs(comment *string, createTime *time.Time, updateTime *time.Time) wargcore.FlagMap {
	now := time.Now()
	commonCreateFlags := wargcore.FlagMap{
		"--comment": flag.New(
			"Comment",
			scalar.String(
				scalar.Default(""),
				scalar.PointerTo(comment),
			),
			flag.Required(),
		),
		"--create-time": flag.New(
			"Create time",
			scalar.New(
				datetime(),
				scalar.Default(now),
				scalar.PointerTo(createTime),
			),
			flag.Required(),
		),
		"--update-time": flag.New(
			"Update time",
			scalar.New(
				datetime(),
				scalar.Default(now),
				scalar.PointerTo(updateTime),
			),
			flag.Required(),
		),
	}
	return commonCreateFlags
}

func commonCreateFlagMap() wargcore.FlagMap {
	now := time.Now()
	commonCreateFlags := wargcore.FlagMap{
		"--comment": flag.New(
			"Comment",
			scalar.String(
				scalar.Default(""),
			),
			flag.Required(),
		),
		"--create-time": flag.New(
			"Create time",
			scalar.New(
				datetime(),
				scalar.Default(now),
			),
			flag.Required(),
		),
		"--update-time": flag.New(
			"Update time",
			scalar.New(
				datetime(),
				scalar.Default(now),
			),
			flag.Required(),
		),
	}
	return commonCreateFlags
}

func commonUpdateFlags() wargcore.FlagMap {

	commonUpdateFlags := wargcore.FlagMap{
		"--comment": flag.New(
			"Comment",
			scalar.String(),
		),
		"--create-time": flag.New(
			"Create time",
			scalar.New(
				datetime(),
			),
		),
		"--new-name": flag.New(
			"New name",
			scalar.String(),
		),
		"--update-time": flag.New(
			"Update time",
			scalar.New(
				datetime(),
				scalar.Default(time.Now()),
			),
			flag.UnsetSentinel("UNSET"),
		),
	}
	return commonUpdateFlags
}

func timeoutFlagMap() wargcore.FlagMap {
	timeoutFlag := wargcore.FlagMap{
		"--timeout": flag.New(
			"Timeout for a run. Use https://pkg.go.dev/time#Duration to build it",
			scalar.Duration(
				scalar.Default(10*time.Minute),
			),
			flag.Required(),
		),
	}
	return timeoutFlag

}

func timeZoneFlagMap() wargcore.FlagMap {
	return wargcore.FlagMap{
		"--timezone": flag.New(
			"Timezone to display dates",
			scalar.String(
				scalar.Default("local"),
				scalar.Choices("local", "utc"),
			),
			flag.Required(),
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

func mustGetCommonCreateArgs(pf wargcore.PassedFlags) commonCreateArgs {
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

func getCommonUpdateArgs(pf wargcore.PassedFlags) commonUpdateArgs {
	return commonUpdateArgs{
		Comment:    ptrFromMap[string](pf, "--comment"),
		CreateTime: ptrFromMap[time.Time](pf, "--create-time"),
		NewName:    ptrFromMap[string](pf, "--new-name"),
		UpdateTime: ptrFromMap[time.Time](pf, "--update-time"),
	}
}

func mustGetEnvNameArg(pf wargcore.PassedFlags) string {
	return pf["--env"].(string)
}

func mustGetMaskArg(pf wargcore.PassedFlags) bool {
	return pf["--mask"].(bool)
}

func mustGetNameArg(pf wargcore.PassedFlags) string {
	return pf["--name"].(string)
}

func mustGetTimeoutArg(pf wargcore.PassedFlags) time.Duration {
	return pf["--timeout"].(time.Duration)
}

func mustGetTimezoneArg(pf wargcore.PassedFlags) string {
	return pf["--timezone"].(string)
}

func mustGetWidthArg(pf wargcore.PassedFlags) int {
	return pf["--width"].(int)
}

// withSetup wraps a cli.Action to read --db-path and --timeout and creates
//   - a context from the timeout
//   - a tracer provider (and sets it globally)
//   - an EnvService ()
func withSetup(
	f func(ctx context.Context, es models.Service, cmdCtx wargcore.Context) error,
) wargcore.Action {
	return func(cmdCtx wargcore.Context) error {

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
	f func(ctx context.Context, es models.Service, cmdCtx wargcore.Context) (*completion.Candidates, error),
) wargcore.CompletionCandidatesFunc {
	return func(cmdCtx wargcore.Context) (*completion.Candidates, error) {

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
func withConfirm(f func(cmdCtx wargcore.Context) error) wargcore.Action {
	return func(cmdCtx wargcore.Context) error {
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
