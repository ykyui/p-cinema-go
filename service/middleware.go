package service

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
)

func PublicApi(n func(*http.Request) (interface{}, int, error)) func(rw http.ResponseWriter, r *http.Request) {
	return func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		result, statusCode, err := n(r)
		if result == nil {
			result = struct{}{}
		}
		rw.WriteHeader(statusCode)
		if err != nil {
			fmt.Println(err)
			resStr, _ := json.Marshal(struct {
				ErrorMsg string `json:"errorMsg"`
			}{err.Error()})
			fmt.Println(string(resStr))
			fmt.Fprint(rw, string(resStr))
		} else {
			resStr, _ := json.Marshal(result)
			fmt.Println(string(resStr))
			fmt.Fprint(rw, string(resStr))
		}
	}
}

func PrivateApi(n func(string, *http.Request) (interface{}, int, error)) func(rw http.ResponseWriter, r *http.Request) {
	return func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		token := r.Header.Get("Authorization")
		username, err := ValidJwt(token)
		if err != nil {
			fmt.Println("token error: ", err)
			rw.WriteHeader(http.StatusUnauthorized)
			resStr, _ := json.Marshal(struct {
				ErrorMsg string `json:"errorMsg"`
			}{err.Error()})
			fmt.Println(string(resStr))
			fmt.Fprint(rw, string(resStr))
			return
		}

		result, statusCode, err := n(username, r)
		if result == nil {
			result = struct{}{}
		}
		rw.WriteHeader(statusCode)
		rw.Header().Set("Content-Type", "application/json")
		if err != nil {
			resStr, _ := json.Marshal(struct {
				ErrorMsg string `json:"errorMsg"`
			}{err.Error()})
			fmt.Fprint(rw, string(resStr))
		} else {
			resStr, _ := json.Marshal(result)
			fmt.Println(string(resStr))
			fmt.Fprint(rw, string(resStr))
		}
	}
}

func HandleImage(next func(r *http.Request) (string, error)) func(rw http.ResponseWriter, r *http.Request) {
	return func(rw http.ResponseWriter, r *http.Request) {
		result, err := next(r)
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
		}
		rw.Header().Set("Content-Type", "image/png")
		dec, err := base64.StdEncoding.DecodeString(result)
		if err != nil {
			panic(err)
		}
		rw.Write(dec)
	}
}
