package main

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.bbkane.com/enventory/app"
	"go.bbkane.com/enventory/models"
	"go.bbkane.com/warg/completion"
	"go.bbkane.com/warg/parseopt"
	"go.bbkane.com/warg/wargcore"
)

// TestMainCompletions tests tab completions
func TestMainCompletions(t *testing.T) {
	t.Parallel()

	dbName := createTempDB(t)

	t.Log("dbFile:", dbName)

	makeComment := func(s string) string {
		return s + " comment"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	es, err := app.NewEnvService(ctx, dbName)
	require.NoError(t, err)
	// create an env
	_, err = es.EnvCreate(ctx, models.EnvCreateArgs{
		Name:       envName01,
		Comment:    makeComment(envName01),
		CreateTime: time.Time{},
		UpdateTime: time.Time{},
	})
	require.NoError(t, err)

	// create a var
	_, err = es.VarCreate(ctx, models.VarCreateArgs{
		EnvName:    envName01,
		Name:       envVarName01,
		Value:      envVarValue01,
		Comment:    makeComment(envVarName01),
		CreateTime: time.Time{},
		UpdateTime: time.Time{},
	})
	require.NoError(t, err)

	// create a  var ref
	_, err = es.VarRefCreate(ctx, models.VarRefCreateArgs{
		EnvName:    envName01,
		Name:       envRefName01,
		Comment:    makeComment(envRefName01),
		CreateTime: time.Time{},
		UpdateTime: time.Time{},
		RefEnvName: envName01,
		RefVarName: envVarName01,
	})
	require.NoError(t, err)

	// We put everything in the db, so we should be able to complete from it.
	app := buildApp()

	completionTests := []struct {
		name               string
		args               []string
		expectedErr        bool
		expectedCandidates *completion.Candidates
	}{
		{
			name:        "envShow",
			args:        []string{"env", "show", "--db-path", dbName, "--name"},
			expectedErr: false,
			expectedCandidates: &completion.Candidates{
				Type: completion.Type_ValuesDescriptions,
				Values: []completion.Candidate{
					{
						Name:        envName01,
						Description: makeComment(envName01),
					},
				},
			},
		},
		{
			name:        "envVarShow",
			args:        []string{"var", "show", "--db-path", dbName, "--env", envName01, "--name"},
			expectedErr: false,
			expectedCandidates: &completion.Candidates{
				Type: completion.Type_ValuesDescriptions,
				Values: []completion.Candidate{
					{
						Name:        envVarName01,
						Description: makeComment(envVarName01),
					},
				},
			},
		},
		{
			name:        "varRefShow",
			args:        []string{"var", "ref", "show", "--db-path", dbName, "--env", envName01, "--name"},
			expectedErr: false,
			expectedCandidates: &completion.Candidates{
				Type: completion.Type_ValuesDescriptions,
				Values: []completion.Candidate{
					{
						Name:        envRefName01,
						Description: makeComment(envRefName01),
					},
				},
			},
		},
		{
			name:        "varRefCreate",
			args:        []string{"var", "ref", "create", "--db-path", dbName, "--ref-env", envName01, "--ref-var"},
			expectedErr: false,
			expectedCandidates: &completion.Candidates{
				Type: completion.Type_ValuesDescriptions,
				Values: []completion.Candidate{
					{
						Name:        envVarName01,
						Description: makeComment(envVarName01),
					},
				},
			},
		},
	}
	for _, tt := range completionTests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)

			// set it up like os.Args
			args := []string{"appName", "--completion-zsh"}
			// add on the test case args
			args = append(args, tt.args...)
			// add on the blank space the shell would add for us
			args = append(args, "")

			actualCandidates, actualErr := app.CompletionCandidates(
				parseopt.Args(args),
				parseopt.LookupEnv(wargcore.LookupMap(nil)),
			)

			if tt.expectedErr {
				require.Error(actualErr)
				return
			} else {
				require.NoError(actualErr)
			}
			require.Equal(tt.expectedCandidates, actualCandidates)
		})
	}

}
