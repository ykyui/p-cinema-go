package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"p-cinema-go/rdbms"
	"strconv"

	"github.com/gorilla/mux"
)

func GetTheatres(r *http.Request) (interface{}, int, error) {
	tx, err := rdbms.TransactionsStart()
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	defer tx.Rollback()
	result, err := (&rdbms.Theatre{}).FindList(tx)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	return result, http.StatusOK, nil
}

func GetTheatreHouses(r *http.Request) (interface{}, int, error) {
	tx, err := rdbms.TransactionsStart()
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	defer tx.Rollback()
	id, _ := strconv.Atoi(r.FormValue("theatreId"))
	result, err := (&rdbms.Theatre{TheatreId: id}).FindHouses(tx)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	return result, http.StatusOK, nil
}

func CreateOrUpdateTheatre(username string, r *http.Request) (interface{}, int, error) {
	var request rdbms.Theatre
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
	if err = request.CreateOrUpdateTheatre(tx); err != nil {
		return nil, http.StatusInternalServerError, err
	}
	if err = tx.Commit(); err != nil {
		return nil, http.StatusInternalServerError, err
	}
	return nil, http.StatusOK, nil
}

func GetTheatresDetail(username string, r *http.Request) (interface{}, int, error) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])
	theatre := rdbms.Theatre{TheatreId: id}
	err := theatre.Find(nil)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	return theatre, http.StatusOK, nil
}

func CreateOrUpdateHouse(username string, r *http.Request) (interface{}, int, error) {
	var request rdbms.House
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
	if err = request.CreateOrUpdateHouse(tx); err != nil {
		return nil, http.StatusInternalServerError, err
	}
	if err = tx.Commit(); err != nil {
		return nil, http.StatusInternalServerError, err
	}
	return nil, http.StatusOK, nil
}

func HouseDetail(username string, r *http.Request) (interface{}, int, error) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])
	house := rdbms.House{HouseId: id}
	if err := house.Find(nil); err != nil {
		return nil, http.StatusInternalServerError, err
	}
	return house, http.StatusOK, nil
}
