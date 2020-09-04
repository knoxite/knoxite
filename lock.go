package knoxite

import (
	"time"
)

type Lock struct {
	Timestamp time.Time `json:"timestamp"` // timestamp the lock was acquired at
	User      string    `json:"user"`      // user@machine that acquired the lock
}

// NewLock creates a new repository lock.
func NewLock() Lock {
	lock := Lock{
		Timestamp: time.Now(),
		User:      "alice@bob",
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

	l, err := repository.backend.LockRepository(b)
	if err != nil {
		return err
	}

	if l != nil {
		return ErrRepositoryLocked
	}

	return nil
}
