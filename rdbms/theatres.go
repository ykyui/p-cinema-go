package rdbms

import (
	"database/sql"
	"fmt"
)

type Theatre struct {
	TheatreId int     `json:"theatreId"`
	Name      string  `json:"name"`
	Houses    []House `json:"houses"`
}

type House struct {
	TheatreId   int    `json:"theatreId"`
	HouseId     int    `json:"houseId"`
	Name        string `json:"name"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	SpecialSeat []Seat `json:"specialSeat"`
}

type Seat struct {
	X        int    `json:"x"`
	Y        int    `json:"y"`
	DisplayX int    `json:"displayX"`
	DisplayY int    `json:"displayY"`
	SeatType string `json:"seatType"`
}

func (t Theatre) FindList(tx *sql.Tx) ([]Theatre, error) {
	stmt, err := tx.Prepare(`select theatres_id, name from theatres`)
	if err != nil {
		return nil, err
	}
	if rows, err := stmt.Query(); err != nil {
		return nil, err
	} else {
		defer rows.Close()
		var (
			theatres_id sql.NullInt64
			name        sql.NullString
		)
		theatres := make([]Theatre, 0)

		for rows.Next() {
			if err = rows.Scan(&theatres_id, &name); err != nil {
				return nil, err
			}
			theatres = append(theatres, Theatre{int(theatres_id.Int64), name.String, nil})
		}
		return theatres, nil
	}
}

func (t *Theatre) Find(tx *sql.Tx) error {
	if tx == nil {
		if _tx, err := db.Begin(); err != nil {
			return err
		} else {
			tx = _tx
			defer _tx.Rollback()
		}

	}
	if t.TheatreId == 0 {
		return nil
	}
	stmt, err := tx.Prepare(`select name from theatres where theatres_id = ?`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	var name sql.NullString
	if err = stmt.QueryRow(t.TheatreId).Scan(&name); err != nil {
		return err
	}
	t.Name = name.String

	houses, err := t.FindHouses(tx)
	if err != nil {
		return err
	}
	t.Houses = houses

	return nil
}

func (t *Theatre) CreateOrUpdateTheatre(tx *sql.Tx) error {
	stmt, err := tx.Prepare(`insert into theatres (theatres_id, name) values (?, ?) on duplicate key update name = values(name)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	if r, err := stmt.Exec(nullInt(t.TheatreId), t.Name); err != nil {
		return err
	} else if t.TheatreId == 0 {
		id, _ := r.LastInsertId()
		t.TheatreId = int(id)
		fmt.Println(id)
	}

	return nil
}

func (t *Theatre) FindHouses(tx *sql.Tx) ([]House, error) {
	stmt, err := tx.Prepare(`select house_id, name from houses where theatre_id = ?`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	if rows, err := stmt.Query(t.TheatreId); err != nil {
		return nil, err
	} else {
		defer rows.Close()

		var (
			house_id sql.NullInt64
			name     sql.NullString
		)
		result := make([]House, 0)
		for rows.Next() {
			if err = rows.Scan(&house_id, &name); err != nil {
				return nil, err
			}
			result = append(result, House{TheatreId: t.TheatreId, HouseId: int(house_id.Int64), Name: name.String})
		}
		return result, nil
	}
}

func (h *House) Find(tx *sql.Tx) error {
	if tx == nil {
		if _tx, err := db.Begin(); err != nil {
			return err
		} else {
			tx = _tx
			defer _tx.Rollback()
		}
	}
	if h.HouseId == 0 {
		return nil
	}
	stmt, err := tx.Prepare(`select name, width, height, absolute_x, absolute_y, seat_type, h.is_active 
	from houses h 
	left join house_seat hs on h.house_id = hs.house_id and hs.is_active = 1 
	where h.house_id = ?`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	if rows, err := stmt.Query(h.HouseId); err != nil {
		return err
	} else {
		defer rows.Close()
		var (
			name       sql.NullString
			width      sql.NullInt64
			height     sql.NullInt64
			absolute_x sql.NullInt64
			absolute_y sql.NullInt64
			seat_type  sql.NullString
			is_active  sql.NullInt64
		)
		for rows.Next() {
			if err = rows.Scan(&name, &width, &height, &absolute_x, &absolute_y, &seat_type, &is_active); err != nil {
				return err
			}
			h.Name = name.String
			h.Width = int(width.Int64)
			h.Height = int(height.Int64)
			if seat_type.Valid {
				h.SpecialSeat = append(h.SpecialSeat, Seat{int(absolute_x.Int64), int(absolute_y.Int64), 0, 0, seat_type.String})
			}
		}
	}

	return nil
}

func (h *House) CreateOrUpdateHouse(tx *sql.Tx) error {
	stmt, err := tx.Prepare(`insert into houses (house_id, theatre_id, name, width, height, is_active) values (?, ?, ?, ?, ?, ?) on duplicate key update name = values(name), width = values(width), height = values(height), is_active = values(is_active)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	stmt_seat, err := tx.Prepare(`insert into house_seat (house_id, absolute_x, absolute_y, seat_type, is_active) values (?, ?, ?, ?, ?) on duplicate key update absolute_x = values(absolute_x), absolute_y = values(absolute_y), seat_type = values(seat_type), is_active = values(is_active)`)
	if err != nil {
		return err
	}
	defer stmt_seat.Close()

	stmt_seat_delete, err := tx.Prepare(`update house_seat set is_active = 0 where house_id = ? and absolute_x = ? and absolute_y = ?`)
	if err != nil {
		return err
	}
	defer stmt_seat_delete.Close()

	if r, err := stmt.Exec(nullInt(h.HouseId), h.TheatreId, h.Name, h.Width, h.Height, 1); err != nil {
		return err
	} else if h.HouseId == 0 {
		id, _ := r.LastInsertId()
		h.HouseId = int(id)

		for _, v := range h.SpecialSeat {
			if _, err := stmt_seat.Exec(h.HouseId, v.X, v.Y, v.SeatType, 1); err != nil {
				return err
			}
		}
	} else {
		oldHouse := &House{HouseId: h.HouseId}
		oldHouse.Find(tx)
		oldHouse.Width = h.Width
		oldHouse.Height = h.Height

		hm := h.SeatMap()
		for i, os := range oldHouse.SeatMap() {
			if s, ok := hm[i]; !ok {
				stmt_seat_delete.Exec(h.HouseId, os.X, os.Y)
			} else {
				if s.SeatType == os.SeatType {
					delete(hm, i)
				}
			}
		}
		for _, v := range hm {
			if _, err := stmt_seat.Exec(h.HouseId, v.X, v.Y, v.SeatType, 1); err != nil {
				return err
			}
		}
	}

	return nil
}

func (h *House) SeatMap() map[int]Seat {
	m := make(map[int]Seat)
	for _, v := range h.SpecialSeat {
		m[v.Y*h.Height+v.X] = v
	}
	return m
}
