package knoxite

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

var verifyTestCases = []struct {
	ErrorFunction          func(dir string) error
	NumberOfExpectedErrors int
	Percentage             int
}{
	{func(dir string) error { return nil }, 0, 100}, // Testing various Percentages
	{func(dir string) error { return nil }, 0, 70},
	{func(dir string) error { return nil }, 0, 30},
	{func(dir string) error { return nil }, 0, 0},
	{func(dir string) error { return nil }, 0, 256},
	{func(dir string) error { return nil }, 0, -256},
	{func(dir string) error {
		// What does happen if all chunks are deleted?
		return os.RemoveAll(filepath.Join(dir, "chunks"))
	}, 2, 100},
	{func(dir string) error {
		// What does happen if all snapshots are deleted?
		return os.RemoveAll(filepath.Join(dir, "snapshots"))
	}, 1, 100},
	{func(dir string) error {
		// What does happen if a specific chunk is deleted?
		layer0, err := ioutil.ReadDir(filepath.Join(dir, "chunks"))
		if err != nil {
			return err
		}
		if len(layer0) == 0 {
			return errors.New("Files expected")
		}

		layer1, err := ioutil.ReadDir(filepath.Join(dir, "chunks", layer0[0].Name()))
		if err != nil {
			return err
		}

		if len(layer1) == 0 {
			return errors.New("Files expected")
		}

		layer2, err := ioutil.ReadDir(filepath.Join(dir, "chunks", layer0[0].Name(), layer1[0].Name()))
		if err != nil {
			return err
		}

		if len(layer2) == 0 {
			return errors.New("Files expected")
		}

		os.Remove(filepath.Join(dir, "chunks", layer0[0].Name(), layer1[0].Name(), layer2[0].Name()))

		return nil

	}, 1, 100},
}

func TestVerifyRepo(t *testing.T) {
	testPassword := "this_is_a_password"

	for _, tt := range verifyTestCases{

		dir, err := ioutil.TempDir("", "knoxite")
		if err != nil {
			t.Errorf("Failed creating temporary dir for repository: %s", err)
			return
		}
		defer os.RemoveAll(dir)
		{
			r, err := NewRepository(dir, testPassword)
			if err != nil {
				t.Errorf("Failed creating repository: %s", err)
				return
			}
			vol, err := NewVolume("test_name", "test_description")
			if err != nil {
				t.Errorf("Failed creating volume: %s", err)
				return
			}
			err = r.AddVolume(vol)
			if err != nil {
				t.Errorf("Failed creating volume: %s", err)
				return
			}
			snapshot, err := NewSnapshot("test_snapshot")
			if err != nil {
				t.Errorf("Failed creating snapshot: %s", err)
				return
			}
			index, err := OpenChunkIndex(&r)
			if err != nil {
				t.Errorf("Failed opening chunk-index: %s", err)
				return
			}

			wd, err := os.Getwd()
			if err != nil {
				t.Errorf("Failed getting working dir: %s", err)
				return
			}

			progress := snapshot.Add(wd, []string{"snapshot_test.go", "snapshot.go"}, []string{}, r, &index, CompressionGZip, EncryptionAES, 1, 0)
			for p := range progress {
				if p.Error != nil {
					t.Errorf("Failed adding to snapshot: %s", p.Error)
				}
			}

			err = snapshot.Save(&r)
			if err != nil {
				t.Errorf("Failed saving snapshot: %s", err)
			}
			err = vol.AddSnapshot(snapshot.ID)
			if err != nil {
				t.Errorf("Failed adding snapshot to volume: %s", err)
			}
			err = r.Save()
			if err != nil {
				t.Errorf("Failed saving volume: %s", err)
				return
			}
			err = index.Save(&r)
			if err != nil {
				t.Errorf("Failed saving chunk-index: %s", err)
				return
			}

		}

		tt.ErrorFunction(dir)
		if err != nil {
			t.Errorf("Failed doing destrutive function: %s", err)
			return
		}

		{
			r, err := OpenRepository(dir, testPassword)
			if err != nil {
				t.Errorf("Failed opening repository: %s", err)
				return
			}

			progress, err := VerifyRepo(r, tt.Percentage)
			if err != nil {
				t.Errorf("Failed to verify snapshot: %s", err)
			}
			errors := make([]error, 0)
			for p := range progress {
				if p.Error != nil {
					errors = append(errors, p.Error)
				}
			}
			if len(errors) != tt.NumberOfExpectedErrors {
				t.Errorf("%d errors where expected but %d occured", tt.NumberOfExpectedErrors, len(errors))
			}
		}

	}

}

func TestVerifyVolume(t *testing.T) {
	testPassword := "this_is_a_password"

	for _, tt := range verifyTestCases {

		dir, err := ioutil.TempDir("", "knoxite")
		if err != nil {
			t.Errorf("Failed creating temporary dir for repository: %s", err)
			return
		}
		defer os.RemoveAll(dir)
		var volumeOriginal *Volume
		{
			r, err := NewRepository(dir, testPassword)
			if err != nil {
				t.Errorf("Failed creating repository: %s", err)
				return
			}
			vol, err := NewVolume("test_name", "test_description")
			if err != nil {
				t.Errorf("Failed creating volume: %s", err)
				return
			}
			err = r.AddVolume(vol)
			if err != nil {
				t.Errorf("Failed creating volume: %s", err)
				return
			}
			snapshot, err := NewSnapshot("test_snapshot")
			if err != nil {
				t.Errorf("Failed creating snapshot: %s", err)
				return
			}
			index, err := OpenChunkIndex(&r)
			if err != nil {
				t.Errorf("Failed opening chunk-index: %s", err)
				return
			}

			wd, err := os.Getwd()
			if err != nil {
				t.Errorf("Failed getting working dir: %s", err)
				return
			}

			progress := snapshot.Add(wd, []string{"snapshot_test.go", "snapshot.go"}, []string{}, r, &index, CompressionGZip, EncryptionAES, 1, 0)
			for p := range progress {
				if p.Error != nil {
					t.Errorf("Failed adding to snapshot: %s", p.Error)
				}
			}

			err = snapshot.Save(&r)
			if err != nil {
				t.Errorf("Failed saving snapshot: %s", err)
			}
			err = vol.AddSnapshot(snapshot.ID)
			if err != nil {
				t.Errorf("Failed adding snapshot to volume: %s", err)
			}
			err = r.Save()
			if err != nil {
				t.Errorf("Failed saving volume: %s", err)
				return
			}
			err = index.Save(&r)
			if err != nil {
				t.Errorf("Failed saving chunk-index: %s", err)
				return
			}

			volumeOriginal = vol
		}

		tt.ErrorFunction(dir)
		if err != nil {
			t.Errorf("Failed doing destrutive function: %s", err)
			return
		}

		{
			r, err := OpenRepository(dir, testPassword)
			if err != nil {
				t.Errorf("Failed opening repository: %s", err)
				return
			}

			progress, err := VerifyVolume(r, volumeOriginal.ID, tt.Percentage)
			if err != nil {
				t.Errorf("Failed to verify snapshot: %s", err)
			}
			errors := make([]error, 0)
			for p := range progress {
				if p.Error != nil {
					errors = append(errors, p.Error)
				}
			}
			if len(errors) != tt.NumberOfExpectedErrors {
				t.Errorf("%d errors where expected but %d occured", tt.NumberOfExpectedErrors, len(errors))
			}
		}

	}

}

func TestVerifySnapshot(t *testing.T) {
	testPassword := "this_is_a_password"

	for _, tt := range verifyTestCases {

		dir, err := ioutil.TempDir("", "knoxite")
		if err != nil {
			t.Errorf("Failed creating temporary dir for repository: %s", err)
			return
		}
		defer os.RemoveAll(dir)
		var snapshotOriginal *Snapshot
		{
			r, err := NewRepository(dir, testPassword)
			if err != nil {
				t.Errorf("Failed creating repository: %s", err)
				return
			}
			vol, err := NewVolume("test_name", "test_description")
			if err != nil {
				t.Errorf("Failed creating volume: %s", err)
				return
			}
			err = r.AddVolume(vol)
			if err != nil {
				t.Errorf("Failed creating volume: %s", err)
				return
			}
			snapshot, err := NewSnapshot("test_snapshot")
			if err != nil {
				t.Errorf("Failed creating snapshot: %s", err)
				return
			}
			index, err := OpenChunkIndex(&r)
			if err != nil {
				t.Errorf("Failed opening chunk-index: %s", err)
				return
			}

			wd, err := os.Getwd()
			if err != nil {
				t.Errorf("Failed getting working dir: %s", err)
				return
			}

			progress := snapshot.Add(wd, []string{"snapshot_test.go", "snapshot.go"}, []string{}, r, &index, CompressionGZip, EncryptionAES, 1, 0)
			for p := range progress {
				if p.Error != nil {
					t.Errorf("Failed adding to snapshot: %s", p.Error)
				}
			}

			err = snapshot.Save(&r)
			if err != nil {
				t.Errorf("Failed saving snapshot: %s", err)
			}
			err = vol.AddSnapshot(snapshot.ID)
			if err != nil {
				t.Errorf("Failed adding snapshot to volume: %s", err)
			}
			err = r.Save()
			if err != nil {
				t.Errorf("Failed saving volume: %s", err)
				return
			}
			err = index.Save(&r)
			if err != nil {
				t.Errorf("Failed saving chunk-index: %s", err)
				return
			}

			snapshotOriginal = snapshot
		}

		tt.ErrorFunction(dir)
		if err != nil {
			t.Errorf("Failed doing destrutive function: %s", err)
			return
		}

		{
			r, err := OpenRepository(dir, testPassword)
			if err != nil {
				t.Errorf("Failed opening repository: %s", err)
				return
			}

			progress, err := VerifySnapshot(r, snapshotOriginal.ID, tt.Percentage)
			if err != nil {
				t.Errorf("Failed to verify snapshot: %s", err)
			}
			errors := make([]error, 0)
			for p := range progress {
				if p.Error != nil {
					errors = append(errors, p.Error)
				}
			}
			if len(errors) != tt.NumberOfExpectedErrors {
				t.Errorf("%d errors where expected but %d occured", tt.NumberOfExpectedErrors, len(errors))
			}
		}

	}
}
