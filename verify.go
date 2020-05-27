/*
 * knoxite
 *     Copyright (c) 2020, Fabian Siegel <fabians1999@gmail.com>
 *
 *   For license see LICENSE
 */

package knoxite




func VerifyArchive(repository Repository, arc Archive) error {
	if arc.Type == File {
		parts := uint(len(arc.Chunks))
		for i := uint(0); i < parts; i++ {
			idx, erri := arc.IndexOfChunk(i)
			if erri != nil {
				return erri
			}

			chunk := arc.Chunks[idx]
			_, errc := loadChunk(repository, arc, chunk)
			if errc != nil {
				return errc
			}

		}
		return nil
	} 
	return nil

}
