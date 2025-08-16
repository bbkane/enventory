package main

import (
	"os"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
	"go.bbkane.com/warg"
)

// abstract some common test functionality to make writing
// tests shorter

const zeroTime = "0001-01-01T00:00:00Z"
const oneTime = "0001-01-01T01:00:00Z"

const envName01 = "envName01"
const envName02 = "envName02"
const envVarName01 = "envVarName01"
const envVarValue01 = "envVarValue01"
const envRefName01 = "envRefName01"

func createTempDB(t *testing.T) string {
	dbFile, err := os.CreateTemp(os.TempDir(), "enventory-test-")
	require.NoError(t, err)
	err = dbFile.Close()
	require.NoError(t, err)

	t.Log("dbFile:", dbFile.Name())
	return dbFile.Name()
}

func envCreateTestCmd(dbPath string, envName string) []string {
	return new(testCmdBuilder).
		Strs("env", "create").
		Name(envName).
		ZeroTimes().
		Finish(dbPath)
}

func envShowTestCmd(dbPath string, envName string) []string {
	return new(testCmdBuilder).
		Strs("env", "show").
		Name(envName).Tz().Mask(false).Finish(dbPath)
}

func varCreateTestCmd(dbPath string, envName string, name string, value string) []string {
	return new(testCmdBuilder).Strs("var", "create").
		EnvName(envName).Name(name).Strs("--value", value).
		ZeroTimes().Finish(dbPath)
}

func varRefCreateTestCmd(dbPath string, envName string, name string, refEnv string, refVar string) []string {
	return new(testCmdBuilder).Strs("var", "ref", "create").
		EnvName(envName).Name(name).Strs("--ref-env", refEnv).
		Strs("--ref-var", refVar).ZeroTimes().Finish(dbPath)
}

type testCmdBuilder struct {
	cmd []string
}

func (tcb *testCmdBuilder) Strs(args ...string) *testCmdBuilder {
	tcb.cmd = append(tcb.cmd, args...)
	return tcb
}

func (tcb *testCmdBuilder) Finish(dbPath string) []string {
	return slices.Concat([]string{"enventory"}, tcb.cmd, []string{"--db-path", dbPath})
}

func (tcb *testCmdBuilder) Name(name string) *testCmdBuilder {
	return tcb.Strs("--name", name)
}

func (tcb *testCmdBuilder) CreateTime(time string) *testCmdBuilder {
	return tcb.Strs("--create-time", time)
}

func (tcb *testCmdBuilder) UpdateTime(time string) *testCmdBuilder {
	return tcb.Strs("--update-time", time)
}

func (tcb *testCmdBuilder) ZeroTimes() *testCmdBuilder {
	return tcb.Strs("--create-time", zeroTime, "--update-time", zeroTime)
}

func (tcb *testCmdBuilder) Comment(comment string) *testCmdBuilder {
	return tcb.Strs("--comment", comment)
}

func (tcb *testCmdBuilder) EnvName(envName string) *testCmdBuilder {
	return tcb.Strs("--env", envName)
}

// Tz sets the timezone to UTC for the command.
func (tcb *testCmdBuilder) Tz() *testCmdBuilder {
	return tcb.Strs("--timezone", "utc")
}

func (tcb *testCmdBuilder) Mask(mask bool) *testCmdBuilder {
	maskStr := "false"
	if mask {
		maskStr = "true"
	}
	return tcb.Strs("--mask", maskStr)
}

func (tcb *testCmdBuilder) Confirm(confirm bool) *testCmdBuilder {
	confirmStr := "false"
	if confirm {
		confirmStr = "true"
	}
	return tcb.Strs("--confirm", confirmStr)
}

type testcase struct {
	name            string
	args            []string
	expectActionErr bool
}

func goldenTest(t *testing.T, tt testcase, updateGolden bool) {
	warg.GoldenTest(
		t,
		warg.GoldenTestArgs{
			App:             buildApp(),
			UpdateGolden:    updateGolden,
			ExpectActionErr: tt.expectActionErr,
		},
		warg.ParseWithArgs(tt.args),
		warg.ParseWithLookupEnv(warg.LookupMap(nil)),
	)
}
