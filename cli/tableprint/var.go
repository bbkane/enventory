package tableprint

import (
	"fmt"

	"go.bbkane.com/enventory/models"
)

func VarShowPrint(c CommonTablePrintArgs, envVar models.Var, envRefs []models.VarRef) {

	switch c.Format {
	case Format_Table:
		t := newKeyValueTable(c.W, c.DesiredMaxWidth)
		createTime := formatTime(envVar.CreateTime, c.Tz)
		updateTime := formatTime(envVar.UpdateTime, c.Tz)
		t.Section(
			newRow("EnvName", envVar.EnvName),
			newRow("Name", envVar.Name),
			newRow("Value", mask(c.Mask, envVar.Value)),
			newRow("Comment", envVar.Comment, skipRowIf(envVar.Comment == "")),
			newRow("CreateTime", createTime),
			newRow("UpdateTime", updateTime, skipRowIf(envVar.CreateTime.Equal(envVar.UpdateTime))),
			newRow("Enabled", fmt.Sprintf("%t", envVar.Enabled), skipRowIf(envVar.Enabled)),
		)
		t.Render()

		if len(envRefs) > 0 {
			fmt.Fprintln(c.W, "EnvRefs")

			t := newKeyValueTable(c.W, c.DesiredMaxWidth)
			for _, e := range envRefs {
				t.Section(
					newRow("EnvName", e.EnvName),
					newRow("Name", e.Name),
					newRow("Comment", e.Comment, skipRowIf(e.Comment == "")),
					newRow("Enabled", fmt.Sprintf("%t", e.Enabled), skipRowIf(e.Enabled)),
				)
			}
			t.Render()
		}
	case Format_ValueOnly:
		fmt.Print(envVar.Value)
	default:
		panic("unexpected format: " + string(c.Format))
	}

}
