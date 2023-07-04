package service

import (
	"errors"
	"p-cinema-go/rdbms"
)

var payRequest = make(chan map[int]*chan bool, 1)
var subscriptPay = make(chan []chan *[]*rdbms.TicketsTransaction, 1)

func init() {
	payRequest <- make(map[int]*chan bool)
	subscriptPay <- make([]chan *[]*rdbms.TicketsTransaction, 0)
}

func PushPay(transactionId int, c *chan bool) error {
	pC := <-payRequest
	defer func() {
		payRequest <- pC
	}()
	pC[transactionId] = c
	subscriptFanout()
	return nil
}

func ApprovePay(transactionId int) error {
	pC := <-payRequest
	defer func() {
		payRequest <- pC
		subscriptFanout()
	}()

	if c, ok := pC[transactionId]; !ok {
		return errors.New("not found")
	} else {
		*c <- true
		delete(pC, transactionId)
	}
	return nil
}

func RejectPay(transactionId int) error {
	pC := <-payRequest
	defer func() {
		payRequest <- pC
		subscriptFanout()
	}()

	if c, ok := pC[transactionId]; !ok {
		return errors.New("not found")
	} else {
		*c <- false
		delete(pC, transactionId)
	}
	return nil
}

func SubscriptPay(c chan *[]*rdbms.TicketsTransaction) error {
	s := <-subscriptPay
	defer func() {
		subscriptPay <- s
	}()
	s = append(s, c)
	return nil
}

func subscriptFanout() error {
	result, err := rdbms.TicketsTransaction{}.WaitingApprovePayment(nil)
	if err != nil {
		return err
	}
	s := <-subscriptPay
	defer func() {
		subscriptPay <- make([]chan *[]*rdbms.TicketsTransaction, 0)
	}()
	for _, v := range s {
		v <- &result
	}
	return nil
}
