// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0
// source: var.sql

package sqlcgen

import (
	"context"
)

const varCreate = `-- name: VarCreate :exec
INSERT INTO var(
    env_id, name, comment, create_time, update_time, value
) VALUES (
    ?     , ?   , ?      , ?          , ?          , ?
)
`

type VarCreateParams struct {
	EnvID      int64
	Name       string
	Comment    string
	CreateTime string
	UpdateTime string
	Value      string
}

func (q *Queries) VarCreate(ctx context.Context, arg VarCreateParams) error {
	_, err := q.db.ExecContext(ctx, varCreate,
		arg.EnvID,
		arg.Name,
		arg.Comment,
		arg.CreateTime,
		arg.UpdateTime,
		arg.Value,
	)
	return err
}

const varDelete = `-- name: VarDelete :execrows
DELETE FROM var WHERE env_id = ? AND name = ?
`

type VarDeleteParams struct {
	EnvID int64
	Name  string
}

func (q *Queries) VarDelete(ctx context.Context, arg VarDeleteParams) (int64, error) {
	result, err := q.db.ExecContext(ctx, varDelete, arg.EnvID, arg.Name)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const varFindByID = `-- name: VarFindByID :one
SELECT env.name AS env_name, var.var_id, var.env_id, var.name, var.comment, var.create_time, var.update_time, var.value
FROM var
JOIN env ON var.env_id = env.env_id
WHERE var.var_id = ?
`

type VarFindByIDRow struct {
	EnvName    string
	VarID      int64
	EnvID      int64
	Name       string
	Comment    string
	CreateTime string
	UpdateTime string
	Value      string
}

func (q *Queries) VarFindByID(ctx context.Context, varID int64) (VarFindByIDRow, error) {
	row := q.db.QueryRowContext(ctx, varFindByID, varID)
	var i VarFindByIDRow
	err := row.Scan(
		&i.EnvName,
		&i.VarID,
		&i.EnvID,
		&i.Name,
		&i.Comment,
		&i.CreateTime,
		&i.UpdateTime,
		&i.Value,
	)
	return i, err
}

const varFindID = `-- name: VarFindID :one
SELECT var_id FROM var WHERE env_id = ? AND name = ?
`

type VarFindIDParams struct {
	EnvID int64
	Name  string
}

func (q *Queries) VarFindID(ctx context.Context, arg VarFindIDParams) (int64, error) {
	row := q.db.QueryRowContext(ctx, varFindID, arg.EnvID, arg.Name)
	var var_id int64
	err := row.Scan(&var_id)
	return var_id, err
}

const varList = `-- name: VarList :many
SELECT var_id, env_id, name, comment, create_time, update_time, value FROM var
WHERE env_id = ?
ORDER BY name ASC
`

func (q *Queries) VarList(ctx context.Context, envID int64) ([]Var, error) {
	rows, err := q.db.QueryContext(ctx, varList, envID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Var
	for rows.Next() {
		var i Var
		if err := rows.Scan(
			&i.VarID,
			&i.EnvID,
			&i.Name,
			&i.Comment,
			&i.CreateTime,
			&i.UpdateTime,
			&i.Value,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const varShow = `-- name: VarShow :one
SELECT var_id, env_id, name, comment, create_time, update_time, value
FROM var
WHERE env_id = ? AND name = ?
`

type VarShowParams struct {
	EnvID int64
	Name  string
}

func (q *Queries) VarShow(ctx context.Context, arg VarShowParams) (Var, error) {
	row := q.db.QueryRowContext(ctx, varShow, arg.EnvID, arg.Name)
	var i Var
	err := row.Scan(
		&i.VarID,
		&i.EnvID,
		&i.Name,
		&i.Comment,
		&i.CreateTime,
		&i.UpdateTime,
		&i.Value,
	)
	return i, err
}

const varUpdate = `-- name: VarUpdate :execrows
UPDATE var SET
    env_id = COALESCE(?1, env_id),
    name = COALESCE(?2, name),
    comment = COALESCE(?3, comment),
    create_time = COALESCE(?4, create_time),
    update_time = COALESCE(?5, update_time),
    value = COALESCE(?6, value)
WHERE var_id = ?7
`

type VarUpdateParams struct {
	EnvID      *int64
	Name       *string
	Comment    *string
	CreateTime *string
	UpdateTime *string
	Value      *string
	VarID      int64
}

func (q *Queries) VarUpdate(ctx context.Context, arg VarUpdateParams) (int64, error) {
	result, err := q.db.ExecContext(ctx, varUpdate,
		arg.EnvID,
		arg.Name,
		arg.Comment,
		arg.CreateTime,
		arg.UpdateTime,
		arg.Value,
		arg.VarID,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
