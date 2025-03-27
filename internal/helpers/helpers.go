package helpers

import (
	"Chirpy/internal/auth"
	"net/http"
	"strings"
)

func CheckApiKey(header http.Header, POLKAKEY string) bool {
	apiKey, err := auth.GetAPIKey(header)
	if err != nil {
		return false
	}
	if apiKey != POLKAKEY {
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
