package main

import (
	"os"
	"testing"
)

func TestBuildApp(t *testing.T) {
	t.Parallel()
	app := buildApp()

	if err := app.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestEnvCreate(t *testing.T) {
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
			name: "03_envList",
			args: new(testCmdBuilder).Strs("env", "list").
				Strs("--timezone", "utc").Finish(dbName),
			expectActionErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			goldenTest(t, tt, updateGolden)
		})
	}
}

func TestEnvDelete(t *testing.T) {
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
			name: "03_envDelete",
			args: new(testCmdBuilder).Strs("env", "delete").
				Confirm(false).Name(envName01).Finish(dbName),
			expectActionErr: false,
		},
		{
			name: "04_envShow",
			args: new(testCmdBuilder).Strs("env", "show").
				Name(envName01).Tz().Mask(false).Finish(dbName),
			expectActionErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			goldenTest(t, tt, updateGolden)
		})
	}
}

func TestEnvUpdate(t *testing.T) {
	t.Parallel()
	updateGolden := os.Getenv("ENVENTORY_TEST_UPDATE_GOLDEN") != ""

	dbName := createTempDB(t)

	tests := []testcase{
		{
			name: "01_envCreate",
			args: new(testCmdBuilder).Strs("env", "create").
				Name(envName01).ZeroTimes().Finish(dbName),
			expectActionErr: false,
		},
		{
			name: "02_envShow",
			args: new(testCmdBuilder).Strs("env", "show").
				Name(envName01).Tz().Mask(false).Finish(dbName),
			expectActionErr: false,
		},
		{
			name: "03_envUpdate",
			args: new(testCmdBuilder).Strs("env", "update").
				Name(envName01).Confirm(false).Strs("--comment", "a comment").
				Strs("--create-time", oneTime).Strs("--new-name", "new_name").
				Strs("--update-time", oneTime).Finish(dbName),
			expectActionErr: false,
		},
		{
			name: "04_envShow",
			args: new(testCmdBuilder).Strs("env", "show").
				Name("new_name").Tz().Mask(false).Finish(dbName),
			expectActionErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			goldenTest(t, tt, updateGolden)
		})
	}
}
