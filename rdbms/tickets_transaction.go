package rdbms

import (
	"database/sql"
	"p-cinema-go/myErr"
	"strings"
)

type TicketsTransaction struct {
	TransactionId  int    `json:"transactionId"`
	FieldId        int    `json:"fieldId"`
	Status         string `json:"status"`
	LastUpdateTime string `json:"lastUpdateTime"`
	Adult          int    `json:"adult"`
	Student        int    `json:"student"`
	Child          int    `json:"child"`
	Disabled       int    `json:"disabled"`
	BoughtSeat     []Seat `json:"boughtSeat"`
}

type TicketsPayment struct {
	TransactionId int `json:"transactionId"`
}

func (t *TicketsTransaction) checkSeatStatus(tx *sql.Tx) error {
	if tx == nil {
		if _tx, err := db.Begin(); err != nil {
			return err
		} else {
			tx = _tx
			defer tx.Rollback()
		}
	}

	stmt_unable_seat, err := tx.Prepare(`select count(*) 
	from field_seat fs, tickets_transaction tr 
	where fs.transaction_id = tr.transaction_id 
	and field_id = ?
	and (
		status = 'success' 
		or (status = 'payment' and tr.transaction_id != ?)
		or (status = 'lock' and AddTime(last_update_time, '00:03:00') > now() and tr.transaction_id != ?) 
	) 
	and (
		0=1 
		` + strings.Repeat(" or (absolute_x = ? and absolute_y = ?) ", len(t.BoughtSeat)) + `
	) for update`)
	if err != nil {
		return err
	}
	defer stmt_unable_seat.Close()

	args := []interface{}{t.FieldId, t.TransactionId, t.TransactionId}
	for _, v := range t.BoughtSeat {
		args = append(args, v.X, v.Y)
	}
	var c sql.NullInt64
	if err = stmt_unable_seat.QueryRow(args...).Scan(&c); err != nil {
		return err
	} else if c.Int64 > 0 {
		return myErr.ErrCantBuy
	}

	return nil
}

func (t *TicketsTransaction) LockSeat(tx *sql.Tx) error {
	stmt_insert_transaction, err := tx.Prepare(`insert into tickets_transaction (status, last_update_time, adult, student, child, disabled) values ('lock', now(), ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt_insert_transaction.Close()

	stmt_insert_transaction_seat, err := tx.Prepare(`insert into field_seat (transaction_id, field_id, absolute_x, absolute_y, display_x, display_y) values (?, ?, ?, ?, ?, ?)` + strings.Repeat(",(?, ?, ?, ?, ?, ?)", len(t.BoughtSeat)-1))
	defer stmt_insert_transaction_seat.Close()

	if err := t.checkSeatStatus(tx); err != nil {
		return err
	}

	if res, err := stmt_insert_transaction.Exec(t.Adult, t.Student, t.Child, t.Disabled); err != nil {
		return err
	} else {
		id, err := res.LastInsertId()
		if err != nil {
			return err
		}
		t.TransactionId = int(id)
	}

	arg := []interface{}{}
	for _, v := range t.BoughtSeat {
		arg = append(arg, t.TransactionId, t.FieldId, v.X, v.Y, v.DisplayX, v.DisplayY)
	}
	if _, err = stmt_insert_transaction_seat.Exec(arg...); err != nil {
		return err
	}

	return nil
}

func (t *TicketsPayment) PaymentLock(tx *sql.Tx, transactionId int) error {
	stmt, err := tx.Prepare(`update tickets_transaction set status = 'payment', last_update_time = now() where transaction_id = ?`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	if _, err := stmt.Exec(transactionId); err != nil {
		return err
	}
	return nil
}

func (t *TicketsPayment) Fail(tx *sql.Tx) error {
	stmt, err := tx.Prepare(`update tickets_transaction set status = 'decline', last_update_time = now() where transaction_id = ?`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	if _, err = stmt.Exec(t.TransactionId); err != nil {
		return err
	}
	return nil
}

func (t *TicketsPayment) Success(tx *sql.Tx) error {
	stmt, err := tx.Prepare(`update tickets_transaction set status = 'success', last_update_time = now() where transaction_id = ?`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	if _, err = stmt.Exec(t.TransactionId); err != nil {
		return err
	}
	return nil
}

func (t TicketsTransaction) WaitingApprovePayment(tx *sql.Tx) ([]*TicketsTransaction, error) {
	if tx == nil {
		if _tx, err := db.Begin(); err != nil {
			return nil, err
		} else {
			tx = _tx
			defer _tx.Rollback()
		}

	}
	stmt, err := tx.Prepare(`select tt.transaction_id, status, last_update_time, adult, student, child, disabled, absolute_x, absolute_y, display_x, display_y
    from tickets_transaction tt, field_seat fs
	where tt.transaction_id = fs.transaction_id
    and status = 'payment'
	order by transaction_id`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	if rows, err := stmt.Query(); err != nil {
		return nil, err
	} else {
		defer rows.Close()
		var (
			transaction_id   sql.NullInt64
			status           sql.NullString
			last_update_time sql.NullString
			adult            sql.NullInt64
			student          sql.NullInt64
			child            sql.NullInt64
			disabled         sql.NullInt64
			absolute_x       sql.NullInt64
			absolute_y       sql.NullInt64
			display_x        sql.NullInt64
			display_y        sql.NullInt64

			result []*TicketsTransaction

			transaction *TicketsTransaction
		)
		for rows.Next() {
			if err = rows.Scan(&transaction_id, &status, &last_update_time, &adult, &student, &child, &disabled, &absolute_x, &absolute_y, &display_x, &display_y); err != nil {
				return nil, err
			}
			if transaction == nil || transaction.TransactionId != int(transaction_id.Int64) {
				transaction = &TicketsTransaction{int(transaction_id.Int64), 0, status.String, last_update_time.String, int(adult.Int64), int(student.Int64), int(child.Int64), int(disabled.Int64), make([]Seat, 0)}
				result = append(result, transaction)
			}
			transaction.BoughtSeat = append(transaction.BoughtSeat, Seat{int(absolute_x.Int64), int(absolute_y.Int64), int(display_x.Int64), int(display_y.Int64), ""})
		}
		return result, nil
	}
}
