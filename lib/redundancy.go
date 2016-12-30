package knoxite

import "github.com/klauspost/reedsolomon"

func redundantData(b []byte, chunks, redundancyChunks int) ([][]byte, error) {
	enc, err := reedsolomon.New(chunks, redundancyChunks)
	if err != nil {
		return [][]byte{}, err
	}

	pars, err := enc.Split(b)
	if err != nil {
		panic(err)
	}

	err = enc.Encode(pars)
	if err != nil {
		return [][]byte{}, err
	}

	return pars, nil
}
