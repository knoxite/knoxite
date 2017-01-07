/*
 * knoxite
 *     Copyright (c) 2016-2017, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE
 */

package knoxite

import "github.com/klauspost/reedsolomon"

func redundantData(b []byte, chunks, redundancyChunks int) ([][]byte, error) {
	enc, err := reedsolomon.New(chunks, redundancyChunks)
	if err != nil {
		return [][]byte{}, err
	}

	pars, err := enc.Split(b)
	if err != nil {
		return [][]byte{}, err
	}

	err = enc.Encode(pars)
	if err != nil {
		return [][]byte{}, err
	}

	return pars, nil
}
