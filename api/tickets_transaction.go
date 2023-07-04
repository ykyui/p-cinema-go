package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"p-cinema-go/rdbms"
	"p-cinema-go/service"
)

func Checkout(r *http.Request) (interface{}, int, error) {
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
		return nil, http.StatusBadRequest, err
	}

	if err = tx.Commit(); err != nil {
		return nil, http.StatusInternalServerError, err
	}
	tokenStr, _ := service.CreateTransactionJwt(request.TransactionId)
	return struct {
		TransactionJwt string `json:"transactionJwt"`
	}{tokenStr}, http.StatusOK, nil
}

func Pay(transactionId int, r *http.Request) (interface{}, int, error) {
	var request rdbms.TicketsPayment
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

	if err = request.PaymentLock(tx, transactionId); err != nil {
		return nil, http.StatusRequestTimeout, err
	}

	if err = tx.Commit(); err != nil {
		return nil, http.StatusInternalServerError, err
	}

	c := make(chan bool, 0)
	defer close(c)
	service.PushPay(transactionId, &c)
	ok := <-c

	return struct {
		Success bool `json:"success"`
	}{ok}, http.StatusOK, nil
}

func WaitingApprovePayment(username string, r *http.Request) (interface{}, int, error) {
	var request struct {
		NumOfAttempt int `json:"numOfAttempt"`
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	err = json.Unmarshal(body, &request)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	if request.NumOfAttempt == 1 {
		result, err := rdbms.TicketsTransaction{}.WaitingApprovePayment(nil)
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}
		return result, http.StatusOK, nil
	}
	c := make(chan *[]*rdbms.TicketsTransaction)
	defer close(c)
	if err = service.SubscriptPay(c); err != nil {
		return nil, http.StatusInternalServerError, err
	}
	return <-c, http.StatusOK, nil
}

func ApprovePayment(username string, r *http.Request) (interface{}, int, error) {
	var request struct {
		TransactionId int `json:"transactionId"`
		Approve       int `json:"approve"`
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

	t := rdbms.TicketsPayment{TransactionId: request.TransactionId}
	if request.Approve == 1 {
		if err = t.Success(tx); err != nil {
			return nil, http.StatusBadRequest, err
		}
	} else {
		if err = t.Fail(tx); err != nil {
			return nil, http.StatusBadRequest, err
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, http.StatusInternalServerError, err
	}
	if request.Approve == 1 {
		service.ApprovePay(request.TransactionId)
	} else {
		service.RejectPay(request.TransactionId)
	}

	return nil, http.StatusOK, nil
}
