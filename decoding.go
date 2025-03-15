package main

import (
	"Chirpy/domain"
	"encoding/json"
	"net/http"
)

func decoding(cfg *apiConfig, w http.ResponseWriter, r *http.Request) (params domain.User) {
	errWrong := domain.ErrStruct{
		Error: "Something went wrong",
	}

	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&params)
	if err != nil {

		cfg.respondWithError(w, http.StatusInternalServerError, errWrong.Error)
		return domain.User{}
	}
	return params
}

func decodingChirp(cfg *apiConfig, w http.ResponseWriter, r *http.Request) (params Chirp) {
	errWrong := domain.ErrStruct{
		Error: "Something went wrong2",
	}
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&params)
	if err != nil {
		cfg.respondWithError(w, http.StatusInternalServerError, errWrong.Error)
		return
	}
	return params
}
