package rdbms

import "database/sql"

type Attachment struct {
	UUID     string
	File     []byte `json:"-"`
	FileType string `json:"fileType"`
}

func (a *Attachment) UploadAttachment() error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`insert into attachment (uuid, file_type) values (?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	if _, err := stmt.Exec(a.UUID, a.FileType); err != nil {
		return err
	}

	return tx.Commit()
}

func (a *Attachment) GetAttachment() error {
	stmt, err := db.Prepare(`select file_type from attachment where uuid = ?`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	var (
		file_type sql.NullString
	)
	if err = stmt.QueryRow(a.UUID).Scan(&file_type); err != nil {
		return err
	}
	a.FileType = file_type.String
	return nil
}
