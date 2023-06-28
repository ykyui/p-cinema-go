package rdbms

import (
	"database/sql"
	"p-cinema-go/service"
	"strings"
)

type Movie struct {
	Id          int      `json:"id"`
	Path        string   `json:"path"`
	Name        string   `json:"name"`
	StartDate   string   `json:"startDate"`
	Cover       string   `json:"cover"`
	Trailer     []string `json:"trailer"`
	Length      int      `json:"length"`
	Ratings     int      `json:"ratings"`
	RatingsDesc string   `json:"ratingsDesc"`
	Desc        string   `json:"desc"`
	Genre       []string `json:"genre"`
	Director    []string `json:"director"`
	Cast        []string `json:"cast"`
	Producers   []string `json:"producers"`
	Writers     []string `json:"writers"`
	Avaliable   int      `json:"avaliable"`
	Promo       int      `json:"promo"`
	Fields      *[]Field `json:"fields"`
}

type Field struct {
	*Theatre `json:"theatre,omitempty"`
	*House   `json:"house,omitempty"`
	*Movie   `json:"movie,omitempty"`
	FieldId  int    `json:"fieldId"`
	ShowDate string `json:"showDate"`
	ShowTime string `json:"showTime"`
	SoldSeat []Seat `json:"soldSeat"`
}

func (m *Movie) GetMovies(tx *sql.Tx) (movies []Movie, err error) {
	movies = make([]Movie, 0)
	if tx == nil {
		_tx, err := db.Begin()
		if err != nil {
			return nil, err
		}
		tx = _tx
		defer tx.Rollback()
	}
	stmt, err := tx.Prepare(`select movie_id, path, name, start_date, cover, length, ratings, desciption, avaliable, promo 
	from movies 
	where (movie_id = ? or ? = 0) 
	and (path = ? or '' = ?) 
	and (name = ? or '' = ?)
	and (start_date <= ? or '' = ?)
	and (avaliable = ? or -1 = ?)`)
	if err != nil {
		return
	}
	defer stmt.Close()

	if rows, err := stmt.Query(m.Id, m.Id, m.Path, m.Path, m.Name, m.Name, m.StartDate, m.StartDate, m.Avaliable, m.Avaliable); err != nil {
		return nil, err
	} else {
		defer rows.Close()
		var (
			movie_id   sql.NullInt64
			path       sql.NullString
			name       sql.NullString
			start_date sql.NullString
			cover      sql.NullString
			length     sql.NullInt64
			ratings    sql.NullInt64
			desc       sql.NullString
			avaliable  sql.NullInt64
			promo      sql.NullInt64
		)
		for rows.Next() {
			if err = rows.Scan(&movie_id, &path, &name, &start_date, &cover, &length, &ratings, &desc, &avaliable, &promo); err != nil {
				return nil, err
			}
			movies = append(movies, Movie{
				int(movie_id.Int64),
				path.String,
				name.String,
				start_date.String,
				cover.String,
				nil,
				int(length.Int64),
				int(ratings.Int64),
				"",
				desc.String,
				nil, nil, nil, nil, nil,
				int(avaliable.Int64),
				int(promo.Int64),
				nil,
			})
		}
	}
	return
}

func (m *Movie) CreateOrUpdateMovie(tx *sql.Tx) error {
	stmt, err := tx.Prepare(`insert into movies (movie_id, path, name, start_date, cover, length, ratings, desciption, avaliable, promo) 
	values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?) 
	on duplicate key update path = values(path), name = values(name), start_date = values(start_date), cover = values(cover), length = values(length), ratings = values(ratings), desciption = values(desciption), avaliable = values(avaliable), promo = values(promo)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	stmt_check_path, err := tx.Prepare(`select count(*) from movies where path = ? and movie_id != ?`)
	if err != nil {
		return err
	}
	defer stmt_check_path.Close()

	var c sql.NullInt64

	if err = stmt_check_path.QueryRow(m.Path, m.Id).Scan(&c); err != nil {
		return err
	} else if c.Int64 != 0 {
		return service.ErrPathDuplicate
	}

	if r, err := stmt.Exec(nullInt(m.Id), m.Path, m.Name, m.StartDate, m.Cover, m.Length, m.Ratings, m.Desc, m.Avaliable, m.Promo); err != nil {
		return err
	} else if m.Id == 0 {
		id, _ := r.LastInsertId()
		m.Id = int(id)
	}

	return nil
}

func (f Field) GetMovieFields(tx *sql.Tx, moviesId []int) (fieldsMap map[int][]Field, fields []Field, err error) {
	if tx == nil {
		_tx, err := db.Begin()
		if err != nil {
			return nil, nil, err
		}
		tx = _tx
		defer tx.Rollback()
	}
	stmt, err := tx.Prepare(`select field_id, theatre_id, house_id, m.movie_id, show_date, show_time, is_active, path, name, start_date, cover, length, ratings, desciption, avaliable, promo
	from fields f, movies m
	where f.movie_id = m.movie_id
	and show_date = ? 
	and theatre_id = ? 
	and (0 = ? or f.movie_id in (0` + strings.Repeat(", ?", len(moviesId)) + `))
	and is_active = 1`)
	if err != nil {
		return nil, nil, err
	}
	defer stmt.Close()

	fieldsMap = make(map[int][]Field)
	args := []interface{}{f.ShowDate, f.Theatre.TheatreId, len(moviesId)}
	for _, v := range moviesId {
		args = append(args, v)
	}
	if rows, err := stmt.Query(args...); err != nil {
		return nil, nil, err
	} else {
		defer rows.Close()

		var (
			field_id   sql.NullInt64
			theatre_id sql.NullInt64
			house_id   sql.NullInt64
			movie_id   sql.NullInt64
			show_date  sql.NullString
			show_time  sql.NullString
			is_active  sql.NullInt64
			path       sql.NullString
			name       sql.NullString
			start_date sql.NullString
			cover      sql.NullString
			length     sql.NullInt64
			ratings    sql.NullInt64
			desciption sql.NullString
			avaliable  sql.NullInt64
			promo      sql.NullInt64
		)
		for rows.Next() {
			if err = rows.Scan(&field_id, &theatre_id, &house_id, &movie_id, &show_date, &show_time, &is_active, &path, &name, &start_date, &cover, &length, &ratings, &desciption, &avaliable, &promo); err != nil {
				return nil, nil, err
			}
			f := Field{
				&Theatre{TheatreId: int(theatre_id.Int64)},
				&House{HouseId: int(house_id.Int64),
					TheatreId: int(theatre_id.Int64)},
				&Movie{Id: int(movie_id.Int64),
					Path:      path.String,
					Name:      name.String,
					StartDate: start_date.String,
					Cover:     cover.String,
					Length:    int(length.Int64),
					Ratings:   int(ratings.Int64),
					Desc:      desciption.String,
					Avaliable: int(avaliable.Int64),
					Promo:     int(promo.Int64),
				}, int(field_id.Int64), show_date.String, show_time.String, nil}
			fieldsMap[int(movie_id.Int64)] = append(fieldsMap[int(movie_id.Int64)], f)
			fields = append(fields, f)
		}
	}
	return
}

func (f Field) UpdateField(fields []Field, tx *sql.Tx) error {
	_, oldFields, err := f.GetMovieFields(tx, []int{})
	if err != nil {
		return err
	}
	stmtDeleteField, err := tx.Prepare(`update fields set is_active = 0 where field_id = ?`)
	if err != nil {
		return err
	}
	defer stmtDeleteField.Close()

	stmtInsertOrUpdateField, err := tx.Prepare(`insert into fields (field_id, theatre_id, house_id, movie_id, show_date, show_time, is_active) values (?, ?, ?, ?, ?, ?, 1) 
	on duplicate key update movie_id = values(movie_id), show_date = values(show_date), show_time = values(show_time), is_active = values(is_active)`)
	if err != nil {
		return err
	}
	defer stmtInsertOrUpdateField.Close()

	for _, o := range oldFields {
		exist := false
		for _, n := range fields {
			if n.FieldId == o.FieldId {
				exist = true
				break
			}
		}
		if !exist {
			if _, err = stmtDeleteField.Exec(o.FieldId); err != nil {
				return err
			}
		}
	}

	for _, v := range fields {
		if _, err := stmtInsertOrUpdateField.Exec(nullInt(v.FieldId), f.Theatre.TheatreId, v.HouseId, v.Id, f.ShowDate, v.ShowTime); err != nil {
			return err
		}
	}

	return nil
}

func (f *Field) SettingPlan() error {
	stmt_field_seat, err := db.Prepare(`select absolute_x, absolute_y, display_x, display_y, status
	from tickets_transaction tt, field_seat fs
	where tt.transaction_id = fs.transaction_id
    and (tt.status = 'success' or (tt.status = 'lock' and AddTime(last_update_time, '00:03:00') > now()))
    and field_id = ?`)
	if err != nil {
		return err
	}
	defer stmt_field_seat.Close()

	stmt_house, err := db.Prepare(`select fields.house_id, name, width, height
	from fields, houses
	where fields.house_id = houses.house_id
	and field_id = ?`)
	if err != nil {
		return err
	}
	defer stmt_house.Close()

	var (
		house_id sql.NullInt64
		name     sql.NullString
		width    sql.NullInt64
		height   sql.NullInt64
	)
	if err = stmt_house.QueryRow(f.FieldId).Scan(&house_id, &name, &width, &height); err != nil {
		return err
	}
	f.House = &House{HouseId: int(house_id.Int64)}

	f.House.Find(nil)
	f.SoldSeat = make([]Seat, 0)
	if rows, err := stmt_field_seat.Query(f.FieldId); err != nil {
		return err
	} else {
		defer rows.Close()
		var (
			absolute_x sql.NullInt64
			absolute_y sql.NullInt64
			display_x  sql.NullInt64
			display_y  sql.NullInt64
			status     sql.NullString
		)
		for rows.Next() {
			if err = rows.Scan(&absolute_x, &absolute_y, &display_x, &display_y, &status); err != nil {
				return err
			}
			f.SoldSeat = append(f.SoldSeat, Seat{int(absolute_x.Int64), int(absolute_y.Int64), int(display_x.Int64), int(display_y.Int64), status.String})
		}
	}
	return nil
}
