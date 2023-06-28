package api

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"p-cinema-go/rdbms"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func UploadAttachment(username string, r *http.Request) (interface{}, int, error) {
	fmt.Println("File Upload Endpoint Hit")

	// Parse our multipart form, 10 << 20 specifies a maximum
	// upload of 10 MB files.
	r.ParseMultipartForm(30 << 20)
	// FormFile returns the first file for the given key `myFile`
	// it also returns the FileHeader so we can get the Filename,
	// the Header and the size of the file
	fmt.Printf("%+v", r)
	file, handler, err := r.FormFile("attachment")
	if err != nil {
		fmt.Println("Error Retrieving the File")
		fmt.Println(err)
		return nil, http.StatusBadRequest, err
	}
	defer file.Close()
	// fmt.Printf("Uploaded File: %+v\n", handler.Filename)
	// fmt.Printf("File Size: %+v\n", handler.Size)
	// fmt.Printf("MIME Header: %+v\n", handler.Header)
	id, _ := uuid.NewRandom()
	f, err := os.Create(fmt.Sprintf("./attachment/%s.%s", id.String(), strings.Split(handler.Filename, ".")[1]))
	if err != nil {
		fmt.Println(err)
		return nil, http.StatusInternalServerError, err
	}
	defer f.Close()

	// read all of the contents of our uploaded file into a
	// byte array
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println(err)
	}
	f.Write(fileBytes)
	a := rdbms.Attachment{UUID: id.String(), FileType: strings.Split(handler.Filename, ".")[1]}
	if err = a.UploadAttachment(); err != nil {
		return nil, http.StatusInternalServerError, err
	}
	// return that we have successfully uploaded our file!
	return struct {
		Id string `json:"id"`
	}{id.String()}, http.StatusOK, nil
}

func AttachmentHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	w.Header().Set("Content-Type", "application/octet-stream")
	a := rdbms.Attachment{UUID: vars["filename"]}
	a.GetAttachment()
	fileBytes, err := ioutil.ReadFile(fmt.Sprintf("./attachment/%s.%s", a.UUID, a.FileType))
	if err != nil {
		fmt.Println(err)
		return
	}
	w.Write(fileBytes)
	// w.Write(a.File)
	w.WriteHeader(http.StatusOK)
}
