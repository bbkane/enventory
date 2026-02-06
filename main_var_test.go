package main

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.bbkane.com/enventory/app"
	"go.bbkane.com/enventory/models"
)

func TestVarCreate(t *testing.T) {
	t.Parallel()
	updateGolden := os.Getenv("ENVENTORY_TEST_UPDATE_GOLDEN") != ""

	dbName := createTempDB(t)

	tests := []testcase{
		{
			name:            "01_envCreate",
			args:            envCreateTestCmd(dbName, envName01),
			expectActionErr: false,
		},
		{
			name:            "02_varCreate",
			args:            varCreateTestCmd(dbName, envName01, varName01, "value"),
			expectActionErr: false,
		},
		{
			name: "03_varShow",
			args: new(testCmdBuilder).Strs("var", "show").
				EnvName(envName01).Name(varName01).Tz().Mask(false).Finish(dbName),
			expectActionErr: false,
		},
		{
			name: "04_envShow",
			args: new(testCmdBuilder).Strs("env", "show").
				Name(envName01).Tz().Mask(false).Finish(dbName),
			expectActionErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			goldenTest(t, tt, updateGolden)
		})
	}
}

func TestVarDelete(t *testing.T) {
	t.Parallel()
	updateGolden := os.Getenv("ENVENTORY_TEST_UPDATE_GOLDEN") != ""

	dbName := createTempDB(t)

	tests := []testcase{

		{
			name:            "01_envCreate",
			args:            envCreateTestCmd(dbName, envName01),
			expectActionErr: false,
		},
		{
			name:            "02_varCreate",
			args:            varCreateTestCmd(dbName, envName01, varName01, "value"),
			expectActionErr: false,
		},
		{
			name: "03_varShow",
			args: new(testCmdBuilder).Strs("var", "show").EnvName(envName01).
				Name(varName01).Tz().Mask(false).Finish(dbName),
			expectActionErr: false,
		},
		{
			name: "04_varDelete",
			args: new(testCmdBuilder).Strs("var", "delete").Confirm(false).
				EnvName(envName01).Name(varName01).Finish(dbName),
			expectActionErr: false,
		},
		{
			name: "05_varShow",
			args: new(testCmdBuilder).Strs("var", "show").EnvName(envName01).
				Name(varName01).Tz().Mask(false).Finish(dbName),
			expectActionErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			goldenTest(t, tt, updateGolden)
		})
	}
}

func TestVarDeleteNonexisting(t *testing.T) {
	t.Parallel()
	updateGolden := os.Getenv("ENVENTORY_TEST_UPDATE_GOLDEN") != ""

	dbName := createTempDB(t)

	tests := []testcase{
		{
			name:            "01_envCreate",
			args:            envCreateTestCmd(dbName, envName01),
			expectActionErr: false,
		},
		// unfortunately, I can't test the output since this just checks that an error occurred
		{
			name: "02_varDeleteNonexisting",
			args: new(testCmdBuilder).Strs("var", "delete").Confirm(false).
				EnvName(envName01).Name("nonexisting").Finish(dbName),
			expectActionErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			goldenTest(t, tt, updateGolden)
		})
	}
}

func TestEnvUniqueNames(t *testing.T) {
	require := require.New(t)
	t.Parallel()

	dbName := createTempDB(t)

	ctx := context.Background()
	service, err := app.NewEnvService(ctx, dbName)
	require.NoError(err)

	// shortcuts!

	createEnv := func(ctx context.Context, name string) error {
		_, err := service.EnvCreate(ctx, models.EnvCreateArgs{Name: name, Comment: "", CreateTime: time.Time{}, UpdateTime: time.Time{}, Enabled: true})
		return err
	}

	createVar := func(ctx context.Context, envName, name string) error {
		_, err := service.VarCreate(ctx, models.VarCreateArgs{EnvName: envName, Name: name, Value: "val", Comment: "", CreateTime: time.Time{}, UpdateTime: time.Time{}, Enabled: true})
		return err
	}
	updateVar := func(ctx context.Context, envName, name, newName string) error {
		err := service.VarUpdate(ctx, envName, name, models.VarUpdateArgs{Comment: nil, CreateTime: nil, EnvName: nil, Name: &newName, UpdateTime: nil, Value: nil, Enabled: nil})
		return err
	}

	createRef := func(ctx context.Context, envName, name, refEnvName, refVarName string) error {
		_, err := service.VarRefCreate(ctx, models.VarRefCreateArgs{EnvName: envName, Name: name, Comment: "", RefEnvName: refEnvName, RefVarName: refVarName, CreateTime: time.Time{}, UpdateTime: time.Time{}, Enabled: true})
		return err
	}

	updateRef := func(ctx context.Context, envName, name, newName string) error {
		err := service.VarRefUpdate(ctx, envName, name, models.VarRefUpdateArgs{RefEnvName: nil, RefVarName: nil, Comment: nil, CreateTime: nil, EnvName: nil, Name: &newName, UpdateTime: nil, Enabled: nil})
		return err
	}

	// Setup. Create env, vars, refs
	err = createEnv(ctx, envName01)
	require.NoError(err)

	err = createVar(ctx, envName01, varName01)
	require.NoError(err)
	err = createVar(ctx, envName01, varName02)
	require.NoError(err)

	err = createRef(ctx, envName01, varRefName01, envName01, varName01)
	require.NoError(err)
	err = createRef(ctx, envName01, varRefName02, envName01, varName02)
	require.NoError(err)

	// Check for conflicts!

	// var create
	err = createVar(ctx, envName01, varName01)
	require.Error(err)
	err = createVar(ctx, envName01, varRefName01)
	require.Error(err)

	// var update
	err = updateVar(ctx, envName01, varName01, varName01)
	require.NoError(err) // update to be the same name should work
	err = updateVar(ctx, envName01, varName01, varName02)
	require.Error(err)
	err = updateVar(ctx, envName01, varName01, varRefName01)
	require.Error(err)

	// ref create
	err = createRef(ctx, envName01, varRefName01, envName01, varName01)
	require.Error(err)
	err = createRef(ctx, envName01, varName01, envName01, varRefName01)
	require.Error(err)

	// ref update
	err = updateRef(ctx, envName01, varRefName01, varRefName01)
	require.NoError(err) // update to be the same name should work
	err = updateRef(ctx, envName01, varRefName01, varName01)
	require.Error(err)
	err = updateRef(ctx, envName01, varRefName01, varRefName02)
	require.Error(err)
}

func TestEnvNonUniqueNames(t *testing.T) {
	t.Parallel()
	updateGolden := os.Getenv("ENVENTORY_TEST_UPDATE_GOLDEN") != ""

	dbName := createTempDB(t)

	tests := []testcase{
		{
			name:            "01_envCreate",
			args:            envCreateTestCmd(dbName, envName01),
			expectActionErr: false,
		},
		{
			name:            "02_varCreate",
			args:            varCreateTestCmd(dbName, envName01, varName01, "value"),
			expectActionErr: false,
		},
		{
			name: "03_varRefCreateSameName",
			args: new(testCmdBuilder).Strs("var", "ref", "create").
				EnvName(envName01).Name(varName01).Strs("--ref-env", envName01).
				Strs("--ref-var", varName01).ZeroTimes().Finish(dbName),
			expectActionErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			goldenTest(t, tt, updateGolden)
		})
	}
}

func TestVarUpdate(t *testing.T) {
	t.Parallel()
	updateGolden := os.Getenv("ENVENTORY_TEST_UPDATE_GOLDEN") != ""

	dbName := createTempDB(t)

	const e1 = "e1"
	const e2 = "e2"
	const v1 = "v1"
	const v2 = "v2"
	const v3 = "v3"

	tests := []testcase{
		// setup
		{
			name:            "01_envCreateE1",
			args:            envCreateTestCmd(dbName, e1),
			expectActionErr: false,
		},
		{
			name:            "02_envCreateE2",
			args:            envCreateTestCmd(dbName, e2),
			expectActionErr: false,
		},
		{
			name:            "03_varCreateE1V1",
			args:            varCreateTestCmd(dbName, e1, v1, "val"),
			expectActionErr: false,
		},
		{
			name:            "04_varCreateE1V2",
			args:            varCreateTestCmd(dbName, e1, v2, "val"),
			expectActionErr: false,
		},
		{
			name:            "05_varCreateE2V2",
			args:            varCreateTestCmd(dbName, e2, v2, "val"),
			expectActionErr: false,
		},
		{
			name:            "06_varCreateE2V3",
			args:            varCreateTestCmd(dbName, e2, v3, "val"),
			expectActionErr: false,
		},
		// test updates
		{
			name:            "07_nameUpdateConflict",
			args:            new(testCmdBuilder).Strs("var", "update").Confirm(false).UpdateTime(oneTime).EnvName(e1).Name(v1).Strs("--new-name", v2).Finish(dbName),
			expectActionErr: true,
		},
		{
			name:            "08_envUpdateConflict",
			args:            new(testCmdBuilder).Strs("var", "update").Confirm(false).UpdateTime(oneTime).EnvName(e1).Name(v2).Strs("--new-env", e2).Finish(dbName),
			expectActionErr: true,
		},
		{
			name:            "09_envUpdateNameUpdateConflict",
			args:            new(testCmdBuilder).Strs("var", "update").Confirm(false).UpdateTime(oneTime).EnvName(e1).Name(v1).Strs("--new-env", e2).Strs("--new-name", v3).Finish(dbName),
			expectActionErr: true,
		},
		{
			name:            "10_emptyUpdate",
			args:            new(testCmdBuilder).Strs("var", "update").Confirm(false).UpdateTime(oneTime).EnvName(e1).Name(v1).Strs("--new-env", e1).Strs("--new-name", v1).Finish(dbName),
			expectActionErr: false,
		},
		{
			name: "11_updateOtherFields",
			args: new(testCmdBuilder).Strs("var", "update").
				Confirm(false).UpdateTime(oneTime).CreateTime(oneTime).EnvName(e1).Name(v1).
				Strs("--new-env", e2, "--new-name", "NEW", "--comment", "NEW", "--value", "NEW").Finish(dbName),
			expectActionErr: false,
		},
		{
			name:            "12_envShowE1",
			args:            envShowTestCmd(dbName, e1),
			expectActionErr: false,
		},
		{
			name:            "13_envShowE2",
			args:            envShowTestCmd(dbName, e2),
			expectActionErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			goldenTest(t, tt, updateGolden)
		})
	}
}
