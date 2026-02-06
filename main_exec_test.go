package main

import (
	"context"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.bbkane.com/enventory/app"
	"go.bbkane.com/enventory/models"
)

//nolint:paralleltest // uses os.Setenv (both for ENVENTORY_EXEC_CONFIG and the exec command)
func TestExec(t *testing.T) {

	if runtime.GOOS == "windows" {
		t.Skip("skipping exec test on windows - no /usr/bin/env")
	}

	require := require.New(t)

	updateGolden := os.Getenv("ENVENTORY_TEST_UPDATE_GOLDEN") != ""

	dbName := createTempDB(t)

	ctx := context.Background()
	service, err := app.NewEnvService(ctx, dbName)
	require.NoError(err)

	_, err = service.EnvCreate(ctx, models.EnvCreateArgs{
		Name:       envName01,
		Comment:    "",
		CreateTime: time.Time{},
		UpdateTime: time.Time{},
		Enabled:    true,
	})
	require.NoError(err)

	_, err = service.VarCreate(ctx, models.VarCreateArgs{
		EnvName:    envName01,
		Name:       "var_from_enventory_env",
		Comment:    "",
		CreateTime: time.Time{},
		UpdateTime: time.Time{},
		Value:      "value_from_enventory_env",
		Enabled:    true,
	})
	require.NoError(err)

	os.Setenv("ENVENTORY_EXEC_CONFIG", "testdata/TestExec/01_exec/exec.yaml")

	// can't use testCmdBuilder because its Finish method appends --db-path after the --
	args := []string{
		"enventory", "exec",
		"--db-path", dbName,
		"--env", envName01,
		"--group", "mygroup",
		"--VAR_WITH_VALUES_COMPLETIONS", "VAR_WITH_VALUES_COMPLETIONS_value",
		"--VAR_WITH_VALUES_DESCRIPTIONS_COMPLETIONS", "VAR_WITH_VALUES_DESCRIPTIONS_COMPLETIONS_value",
		"--",
		"/bin/bash", "--noprofile", "--norc", "--restricted",
		"-c", "echo -n $var_from_enventory_env $GROUP_VAR1 $VAR_WITH_VALUES_COMPLETIONS $VAR_WITH_VALUES_DESCRIPTIONS_COMPLETIONS",
	}

	tt := testcase{
		name:            "01_exec",
		args:            args,
		expectActionErr: false,
	}

	t.Run(tt.name, func(t *testing.T) {
		goldenTest(t, tt, updateGolden)
	})

	// cleanup!
	os.Unsetenv("ENVENTORY_EXEC_CONFIG")
	os.Unsetenv("var_from_enventory_env")
	os.Unsetenv("GROUP_VAR1")
	os.Unsetenv("GROUP_VAR2")
	os.Unsetenv("VAR_WITH_VALUES_COMPLETIONS")
	os.Unsetenv("VAR_WITH_VALUES_DESCRIPTIONS_COMPLETIONS")
}
