package decoding

import (
	"Chirpy/config"
	"Chirpy/domain"
	"Chirpy/respond"
	"encoding/json"
	"net/http"
)

func Decoding(cfg *config.Config, w http.ResponseWriter, r *http.Request) (params domain.User) {
	errWrong := domain.ErrStruct{
		Error: "Something went wrong",
	}

	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&params)
	if err != nil {

		respond.RespondWithError(w, http.StatusInternalServerError, errWrong.Error)
		return domain.User{}
	}
	return params
}

func DecodingChirp(cfg *config.Config, w http.ResponseWriter, r *http.Request) (params domain.Chirp) {
	errWrong := domain.ErrStruct{
		Error: "Something went wrong2",
	}
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&params)
	if err != nil {
		respond.RespondWithError(w, http.StatusInternalServerError, errWrong.Error)
		return
	}
	return params
}
