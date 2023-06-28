package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"p-cinema-go/rdbms"
	"p-cinema-go/service"
)

func PurchaseTickets(r *http.Request) (interface{}, int, error) {
	var request rdbms.TicketsTransaction
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

	if err = request.LockSeat(tx); err != nil {
		return nil, http.StatusInternalServerError, err
	}

	if err = tx.Commit(); err != nil {
		return nil, http.StatusInternalServerError, err
	}
	tokenStr, _ := service.CreateTransactionJwt(request.TransactionId)
	return struct {
		TransactionJwt string `json:"transactionJwt"`
	}{tokenStr}, http.StatusOK, nil
}

func PurchaseProcess(transactionId int, r *http.Request) (interface{}, int, error) {
	var request rdbms.TicketsPayment
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	err = json.Unmarshal(body, &request)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	request.TransactionId = transactionId

	tx, err := rdbms.TransactionsStart()
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	defer tx.Rollback()

	if err = request.PaymentLock(tx); err != nil {
		return nil, http.StatusRequestTimeout, err
	}

	if err = request.Success(tx); err != nil {
		return nil, http.StatusBadRequest, err
	}

	if err = tx.Commit(); err != nil {
		return nil, http.StatusInternalServerError, err
	}
	return nil, http.StatusOK, nil
}
