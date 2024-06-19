package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record Not Found (data/models file)")
	ErrEditConflict   = errors.New("ErrEditConfflict(data/models file)")
)

type Models struct {
	Movies      MovieModel
	Users       UserModel
	Tokens      TokenModel
	Permissions PermissionModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Movies:      MovieModel{Db: db},
		Users:       UserModel{DB: db},
		Tokens:      TokenModel{DB: db},
		Permissions: PermissionModel{DB: db},
	}
}
