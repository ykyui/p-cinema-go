package admin

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"p-cinema-go/service"
)

func Login(r *http.Request) (interface{}, int, error) {
	var request struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	err = json.Unmarshal(body, &request)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	token, err := service.CreateJwt(request.Username, 1)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	return struct {
		Token string `json:"token"`
	}{token}, http.StatusOK, nil
}
