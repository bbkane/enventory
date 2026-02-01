package main

import (
	"os"
	"testing"

	"go.bbkane.com/enventory/cli"
	"go.bbkane.com/warg"
	"go.bbkane.com/warg/metadata"
)

func TestShellZshExportNoEnvNoProblem(t *testing.T) {
	t.Parallel()
	updateGolden := os.Getenv("ENVENTORY_TEST_UPDATE_GOLDEN") != ""

	dbName := createTempDB(t)

	tests := []testcase{
		{
			name: "01_envExport",
			args: new(testCmdBuilder).Strs("shell", "zsh", "export").
				EnvName("non-existent-env").Finish(dbName),
			expectActionErr: true,
		},
		{
			name: "01_envPrintScriptNoEnvNoProblem",
			args: new(testCmdBuilder).Strs("shell", "zsh", "export").
				EnvName("non-existent-env").Strs("--no-env-no-problem", "true").
				Finish(dbName),
			expectActionErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			goldenTest(t, tt, updateGolden)
		})
	}
}

func TestShellZshExport(t *testing.T) {
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
			args:            varCreateTestCmd(dbName, envName01, varName01, varValue01),
			expectActionErr: false,
		},
		{
			name: "03_export",
			args: new(testCmdBuilder).Strs("shell", "zsh", "export").
				EnvName(envName01).Finish(dbName),
			expectActionErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			goldenTest(t, tt, updateGolden)
		})
	}
}

func TestShellZshUnexport(t *testing.T) {
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
			args:            varCreateTestCmd(dbName, envName01, varName01, varValue01),
			expectActionErr: false,
		},
		{
			name: "03_unexport",
			args: new(testCmdBuilder).Strs("shell", "zsh", "unexport").
				EnvName(envName01).Finish(dbName),
			expectActionErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			goldenTest(t, tt, updateGolden)
		})
	}
}

func TestShellZshChdir(t *testing.T) {
	t.Parallel()
	updateGolden := os.Getenv("ENVENTORY_TEST_UPDATE_GOLDEN") != ""

	dbName := createTempDB(t)

	tests := []testcase{
		{
			name:            "01_oldEnvCreate",
			args:            envCreateTestCmd(dbName, "old"),
			expectActionErr: false,
		},
		{
			name:            "02_oldVarCreate",
			args:            varCreateTestCmd(dbName, "old", "ov1", "ov1val - key with same val in new"),
			expectActionErr: false,
		},
		{
			name:            "03_oldVarRefCreate",
			args:            varRefCreateTestCmd(dbName, "old", "or1", "old", "ov1"),
			expectActionErr: false,
		},
		{
			name:            "05_newEnvCreate",
			args:            envCreateTestCmd(dbName, "new"),
			expectActionErr: false,
		},
		{
			name:            "06_newVarCreate",
			args:            varCreateTestCmd(dbName, "new", "nv1", "nv1val"),
			expectActionErr: false,
		},
		{
			name:            "07_newVarRefCreate",
			args:            varRefCreateTestCmd(dbName, "new", "nr1", "new", "nv1"),
			expectActionErr: false,
		},
		{
			name:            "08_newVarCreateSameName",
			args:            varCreateTestCmd(dbName, "new", "ov1", "ov1val-in-new-env"),
			expectActionErr: false,
		},
		{
			name:            "09_oldEnvShow",
			args:            envShowTestCmd(dbName, "old"),
			expectActionErr: false,
		},
		{
			name:            "10_newEnvShow",
			args:            envShowTestCmd(dbName, "new"),
			expectActionErr: false,
		},
		{
			name: "11_chdir",
			args: new(testCmdBuilder).Strs("shell", "zsh", "chdir").
				Strs("--old", "old", "--new", "new").Finish(dbName),
			expectActionErr: false,
		},
	}

	md := metadata.New(cli.CustomLookupEnvFuncKey{}, cli.LookupMap(map[string]string{
		"ov1": "ov1val",
		"or1": "ov1val",
		"nr1": "nv1val",
	}))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			warg.GoldenTest(
				t,
				warg.GoldenTestArgs{
					App:             buildApp(),
					UpdateGolden:    updateGolden,
					ExpectActionErr: tt.expectActionErr,
				},
				warg.ParseWithArgs(tt.args),
				warg.ParseWithLookupEnv(warg.LookupMap(nil)),
				warg.ParseWithMetadata(md),
			)
		})
	}
}
