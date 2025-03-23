package helpers

import (
	"Chirpy/config"
	"Chirpy/internal/auth"
	"Chirpy/respond"
	"net/http"
	"strings"
)

func CheckApiKey(Cfg *config.Config, w http.ResponseWriter, r *http.Request) bool {
	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		respond.RespondWithError(w, http.StatusUnauthorized, "Invalid Authorization header")
		return false
	}
	if apiKey != Cfg.PolkaKey {
		respond.RespondWithError(w, http.StatusUnauthorized, "Invalid API key")
		return false
	}
	return true
}

func CleanBody(body string) (bodyResp string) {
	bodySlice := strings.Split(body, " ")
	var bodySliceResp []string
	for _, v := range bodySlice {
		l := strings.ToLower(v)
		if l != "kerfuffle" && l != "sharbert" && l != "fornax" {
			bodySliceResp = append(bodySliceResp, v)
		} else {
			bodySliceResp = append(bodySliceResp, "****")
		}
	}
	bodyResp = strings.Join(bodySliceResp, " ")
	return
}
