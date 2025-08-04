package tableprint

import (
	"fmt"

	"go.bbkane.com/enventory/models"
)

func EnvList(c CommonTablePrintArgs, envs []models.Env) {
	if len(envs) > 0 {
		t := newKeyValueTable(c.W, c.DesiredMaxWidth)
		for _, e := range envs {
			createTime := formatTime(e.CreateTime, c.Tz)
			updateTime := formatTime(e.UpdateTime, c.Tz)
			t.Section(
				newRow("Name", e.Name),
				newRow("Comment", e.Comment, skipRowIf(e.Comment == "")),
				newRow("CreateTime", createTime),
				newRow("UpdateTime", updateTime, skipRowIf(e.CreateTime.Equal(e.UpdateTime))),
			)
		}
		t.Render()
	} else {
		fmt.Fprintln(c.W, "no envs found")
	}
}

func EnvShowRun(
	c CommonTablePrintArgs,
	env models.Env,
	localvars []models.Var,
	refs []models.VarRef,
	referencedVars []models.Var,
) {
	switch c.Format {
	case Format_Table:
		fmt.Fprintln(c.W, "Env")

		t := newKeyValueTable(c.W, c.DesiredMaxWidth)
		createTime := formatTime(env.CreateTime, c.Tz)
		updateTime := formatTime(env.UpdateTime, c.Tz)
		t.Section(
			newRow("Name", env.Name),
			newRow("Comment", env.Comment, skipRowIf(env.Comment == "")),
			newRow("CreateTime", createTime),
			newRow("UpdateTime", updateTime, skipRowIf(env.CreateTime.Equal(env.UpdateTime))),
		)
		t.Render()

		if len(localvars) > 0 {
			fmt.Fprintln(c.W, "Vars")

			// TODO: need to redesign this to dynamically compute the length of the longest key because some keys can be skipped (like "Comment" here)
			t := newKeyValueTable(c.W, c.DesiredMaxWidth)
			for _, e := range localvars {
				t.Section(
					newRow("Name", e.Name),
					newRow("Value", mask(c.Mask, e.Value)),
					newRow("Comment", e.Comment, skipRowIf(e.Comment == "")),
				)
			}
			t.Render()
		}

		if len(refs) > 0 {
			fmt.Fprintln(c.W, "Refs")
			t := newKeyValueTable(c.W, c.DesiredMaxWidth)

			for i := range len(refs) {
				t.Section(
					newRow("Name", refs[i].Name),
					newRow("RefEnvName", referencedVars[i].EnvName),
					newRow("RefVarName", referencedVars[i].Name),
					newRow("RefVarValue", mask(c.Mask, referencedVars[i].Value)),
					newRow("Comment", refs[i].Comment, skipRowIf(refs[i].Comment == "")),
				)
			}
			t.Render()

		}
	default:
		panic("unexpected format: " + string(c.Format))
	}
}
