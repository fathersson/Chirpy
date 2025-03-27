package decoding

import (
	"Chirpy/internal/domain"
	"encoding/json"
)

func DecodingUser(body []byte) (params domain.User, err error) {
	err = json.Unmarshal(body, &params)
	if err != nil {
		return domain.User{}, err
	}
	return params, nil
}

func DecodingChirp(body []byte) (params domain.Chirp, err error) {
	err = json.Unmarshal(body, &params)
	if err != nil {
		return domain.Chirp{}, err
	}
	return params, nil
}
