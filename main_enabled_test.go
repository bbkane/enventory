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

func TestEnvCreateDisabled(t *testing.T) {
	t.Parallel()
	updateGolden := os.Getenv("ENVENTORY_TEST_UPDATE_GOLDEN") != ""

	dbName := createTempDB(t)

	tests := []testcase{
		{
			name: "01_envCreateDisabled",
			args: new(testCmdBuilder).Strs("env", "create").
				Name(envName01).ZeroTimes().Enabled(false).Finish(dbName),
			expectActionErr: false,
		},
		{
			name: "02_envShow",
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

func TestEnvUpdateEnabled(t *testing.T) {
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
			name: "02_envShow",
			args: new(testCmdBuilder).Strs("env", "show").
				Name(envName01).Tz().Mask(false).Finish(dbName),
			expectActionErr: false,
		},
		{
			name: "03_envUpdateDisable",
			args: new(testCmdBuilder).Strs("env", "update").
				Name(envName01).Confirm(false).Enabled(false).
				Strs("--update-time", "UNSET").Finish(dbName),
			expectActionErr: false,
		},
		{
			name: "04_envShowDisabled",
			args: new(testCmdBuilder).Strs("env", "show").
				Name(envName01).Tz().Mask(false).Finish(dbName),
			expectActionErr: false,
		},
		{
			name: "05_envUpdateEnable",
			args: new(testCmdBuilder).Strs("env", "update").
				Name(envName01).Confirm(false).Enabled(true).
				Strs("--update-time", "UNSET").Finish(dbName),
			expectActionErr: false,
		},
		{
			name: "06_envShowEnabled",
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

func TestVarCreateDisabled(t *testing.T) {
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
			name: "02_varCreateDisabled",
			args: new(testCmdBuilder).Strs("var", "create").
				EnvName(envName01).Name(varName01).Strs("--value", "value").
				ZeroTimes().Enabled(false).Finish(dbName),
			expectActionErr: false,
		},
		{
			name: "03_varShow",
			args: new(testCmdBuilder).Strs("var", "show").
				EnvName(envName01).Name(varName01).Tz().Mask(false).Finish(dbName),
			expectActionErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			goldenTest(t, tt, updateGolden)
		})
	}
}

func TestVarUpdateEnabled(t *testing.T) {
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
			name: "04_varUpdateDisable",
			args: new(testCmdBuilder).Strs("var", "update").
				EnvName(envName01).Name(varName01).Confirm(false).Enabled(false).
				Strs("--update-time", "UNSET").Finish(dbName),
			expectActionErr: false,
		},
		{
			name: "05_varShowDisabled",
			args: new(testCmdBuilder).Strs("var", "show").
				EnvName(envName01).Name(varName01).Tz().Mask(false).Finish(dbName),
			expectActionErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			goldenTest(t, tt, updateGolden)
		})
	}
}

func TestVarRefCreateDisabled(t *testing.T) {
	t.Parallel()
	updateGolden := os.Getenv("ENVENTORY_TEST_UPDATE_GOLDEN") != ""

	dbName := createTempDB(t)

	tests := []testcase{
		{
			name:            "01_envCreate01",
			args:            envCreateTestCmd(dbName, envName01),
			expectActionErr: false,
		},
		{
			name:            "02_varCreate",
			args:            varCreateTestCmd(dbName, envName01, varName01, "val01"),
			expectActionErr: false,
		},
		{
			name:            "03_envCreate02",
			args:            envCreateTestCmd(dbName, envName02),
			expectActionErr: false,
		},
		{
			name: "04_varRefCreateDisabled",
			args: new(testCmdBuilder).Strs("var", "ref", "create").
				EnvName(envName02).Name(varRefName01).ZeroTimes().Enabled(false).
				Strs("--ref-env", envName01).
				Strs("--ref-var", varName01).Finish(dbName),
			expectActionErr: false,
		},
		{
			name: "05_varRefShow",
			args: new(testCmdBuilder).Strs("var", "ref", "show").
				EnvName(envName02).Name(varRefName01).Tz().Mask(false).Finish(dbName),
			expectActionErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			goldenTest(t, tt, updateGolden)
		})
	}
}

func TestVarRefUpdateEnabled(t *testing.T) {
	t.Parallel()
	updateGolden := os.Getenv("ENVENTORY_TEST_UPDATE_GOLDEN") != ""

	dbName := createTempDB(t)

	tests := []testcase{
		{
			name:            "01_envCreate01",
			args:            envCreateTestCmd(dbName, envName01),
			expectActionErr: false,
		},
		{
			name:            "02_varCreate",
			args:            varCreateTestCmd(dbName, envName01, varName01, "val01"),
			expectActionErr: false,
		},
		{
			name:            "03_envCreate02",
			args:            envCreateTestCmd(dbName, envName02),
			expectActionErr: false,
		},
		{
			name:            "04_varRefCreate",
			args:            varRefCreateTestCmd(dbName, envName02, varRefName01, envName01, varName01),
			expectActionErr: false,
		},
		{
			name: "05_varRefShow",
			args: new(testCmdBuilder).Strs("var", "ref", "show").
				EnvName(envName02).Name(varRefName01).Tz().Mask(false).Finish(dbName),
			expectActionErr: false,
		},
		{
			name: "06_varRefUpdateDisable",
			args: new(testCmdBuilder).Strs("var", "ref", "update").
				EnvName(envName02).Name(varRefName01).Confirm(false).Enabled(false).
				Strs("--update-time", "UNSET").Finish(dbName),
			expectActionErr: false,
		},
		{
			name: "07_varRefShowDisabled",
			args: new(testCmdBuilder).Strs("var", "ref", "show").
				EnvName(envName02).Name(varRefName01).Tz().Mask(false).Finish(dbName),
			expectActionErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			goldenTest(t, tt, updateGolden)
		})
	}
}

func TestEnvListWithExprEnabled(t *testing.T) {
	require := require.New(t)
	t.Parallel()

	dbName := createTempDB(t)

	ctx := context.Background()
	service, err := app.NewEnvService(ctx, dbName)
	require.NoError(err)

	// Create an enabled env
	_, err = service.EnvCreate(ctx, models.EnvCreateArgs{
		Name:       "enabledenv",
		Comment:    "",
		CreateTime: time.Time{},
		UpdateTime: time.Time{},
		Enabled:    true,
	})
	require.NoError(err)

	// Create a disabled env
	_, err = service.EnvCreate(ctx, models.EnvCreateArgs{
		Name:       "disabledenv",
		Comment:    "",
		CreateTime: time.Time{},
		UpdateTime: time.Time{},
		Enabled:    false,
	})
	require.NoError(err)

	// Filter for disabled envs only
	query := "filter(Envs, .Enabled == false)"
	actualEnvs, err := service.EnvList(ctx, models.EnvListArgs{
		Expr: &query,
	})
	require.NoError(err)
	require.Len(actualEnvs, 1)
	require.Equal("disabledenv", actualEnvs[0].Name)
	require.False(actualEnvs[0].Enabled)

	// Filter for enabled envs only
	query2 := "filter(Envs, .Enabled == true)"
	actualEnvs2, err := service.EnvList(ctx, models.EnvListArgs{
		Expr: &query2,
	})
	require.NoError(err)
	require.Len(actualEnvs2, 1)
	require.Equal("enabledenv", actualEnvs2[0].Name)
	require.True(actualEnvs2[0].Enabled)
}
