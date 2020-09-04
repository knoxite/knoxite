package knoxite

import (
	"fmt"
	"time"

	"github.com/dustin/go-humanize"
)

type Lock struct {
	Timestamp time.Time `json:"timestamp"` // timestamp the lock was acquired at
	User      string    `json:"user"`      // user that acquired the lock
	Host      string    `json:"host"`      // host that the lock got acquired on
}

type RepositoryLockedError struct {
	ExistingLock Lock
}

func (e *RepositoryLockedError) Error() string {
	return fmt.Sprintf("Repository has been locked by another user: %s@%s (%s)",
		e.ExistingLock.User,
		e.ExistingLock.Host,
		humanize.Time(e.ExistingLock.Timestamp),
	)
}

// NewLock creates a new repository lock.
func NewLock() Lock {
	lock := Lock{
		Timestamp: time.Now(),
		User:      CurrentUser(),
		Host:      CurrentHost(),
	}

	return lock
}

// Save writes a lock.
func (lock Lock) Save(repository *Repository) error {
	pipe, err := NewEncodingPipeline(CompressionLZMA, EncryptionAES, repository.Key)
	if err != nil {
		return err
	}
	b, err := pipe.Encode(lock)
	if err != nil {
		return err
	}

	lb, err := repository.backend.LockRepository(b)
	if err != nil {
		return err
	}

	if lb != nil {
		var l Lock
		pipe, err := NewDecodingPipeline(CompressionLZMA, EncryptionAES, repository.Key)
		if err != nil {
			return err
		}
		err = pipe.Decode(lb, &l)
		if err != nil {
			return err
		}

		return &RepositoryLockedError{
			ExistingLock: l,
		}
	}

	return nil
}
