package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"p-cinema-go/rdbms"
	"p-cinema-go/service"
	"strconv"

	"github.com/gorilla/mux"
)

func GetAvaliableMovies(r *http.Request) (interface{}, int, error) {
	movies, err := (&rdbms.Movie{StartDate: r.FormValue("date"), Avaliable: 1}).GetMovies(nil)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	return movies, http.StatusOK, nil
}

func GetMovieDetail(r *http.Request) (interface{}, int, error) {
	vars := mux.Vars(r)
	movies, err := (&rdbms.Movie{Path: vars["moviePath"], Avaliable: -1}).GetMovies(nil)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	if len(movies) == 1 {
		return movies[0], http.StatusOK, nil
	}
	return nil, http.StatusBadRequest, nil
}

func SearchMovie(r *http.Request) (interface{}, int, error) {
	date := r.FormValue("date")
	movieId, _ := strconv.Atoi(r.FormValue("movieId"))
	theatreId, err := strconv.Atoi(r.FormValue("theatreId"))
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	movies, err := (&rdbms.Movie{Id: movieId, Avaliable: -1, StartDate: date}).GetMovies(nil)
	movieIds := []int{}
	for _, v := range movies {
		movieIds = append(movieIds, v.Id)
	}
	if date != "" {
		fields, _, _ := (&rdbms.Field{ShowDate: date, Theatre: &rdbms.Theatre{TheatreId: theatreId}}).GetMovieFields(nil, movieIds)
		for i, m := range movies {
			f := fields[m.Id]
			movies[i].Fields = &f
		}
	}
	return movies, http.StatusOK, nil
}

func GetTheatreField(r *http.Request) (interface{}, int, error) {
	date := r.FormValue("date")
	theatreId, err := strconv.Atoi(r.FormValue("theatreId"))
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	_, fields, err := (&rdbms.Field{ShowDate: date, Theatre: &rdbms.Theatre{TheatreId: theatreId}}).GetMovieFields(nil, []int{})
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	return fields, http.StatusOK, nil
}

func GetMovieDetailById(username string, r *http.Request) (interface{}, int, error) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])
	movies, err := (&rdbms.Movie{Id: id, Avaliable: -1}).GetMovies(nil)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	if len(movies) == 1 {
		return movies[0], http.StatusOK, nil
	}
	return nil, http.StatusBadRequest, nil
}

func CreateOrUpdateMovie(username string, r *http.Request) (interface{}, int, error) {
	var request rdbms.Movie
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	err = json.Unmarshal(body, &request)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	tx, err := rdbms.TransactionsStart()
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	defer tx.Rollback()
	if err = request.CreateOrUpdateMovie(tx); err == service.ErrPathDuplicate {
		return nil, service.ErrPathDuplicateResponseStatusCode, err
	} else if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	return nil, http.StatusOK, tx.Commit()
}
func CreateOrUpdateField(username string, r *http.Request) (interface{}, int, error) {
	var request struct {
		Date      string        `json:"date"`
		TheatreId int           `json:"theatreId"`
		Fields    []rdbms.Field `json:"fields"`
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	err = json.Unmarshal(body, &request)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	tx, err := rdbms.TransactionsStart()
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	defer tx.Rollback()
	f := rdbms.Field{ShowDate: request.Date, Theatre: &rdbms.Theatre{TheatreId: request.TheatreId}}
	if err = f.UpdateField(request.Fields, tx); err != nil {
		return nil, http.StatusInternalServerError, err
	}
	if err = tx.Commit(); err != nil {
		return nil, http.StatusInternalServerError, err
	}
	return nil, http.StatusOK, nil
}

func GetFieldSettingPlan(r *http.Request) (interface{}, int, error) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	field := rdbms.Field{FieldId: id, Theatre: nil}
	err = field.SettingPlan()
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	return field, http.StatusOK, nil
}
