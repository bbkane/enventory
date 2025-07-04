package main

import (
	"os"
	"testing"
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
			args:            varCreateTestCmd(dbName, envName01, envVarName01, envVarValue01),
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
			args:            varCreateTestCmd(dbName, envName01, envVarName01, envVarValue01),
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
