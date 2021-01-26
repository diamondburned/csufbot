// Package badger provides the BadgerDB backend for storing session data.
package badger

import (
	"bytes"
	"encoding/gob"
	"log"
	"time"

	"github.com/dgraph-io/badger/v3"
	"github.com/dgraph-io/badger/v3/options"
	"github.com/diamondburned/csufbot/internal/session"
	"github.com/pkg/errors"
)

type Store struct {
	db *badger.DB
}

var _ session.Storer = (*Store)(nil)

// New opens the Badger store from the given path with optimized options.
func New(path string) (*Store, error) {
	opts := badger.LSMOnlyOptions(path)
	opts = opts.WithLoggingLevel(badger.WARNING)
	opts.ChecksumVerificationMode = options.OnTableRead
	opts.CompactL0OnClose = true
	opts.Compression = options.Snappy
	opts.DetectConflicts = false

	db, err := badger.Open(opts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open db")
	}

	return FromDB(db), nil
}

// FromDB creates a new Badger store from an existing database.
func FromDB(db *badger.DB) *Store {
	return &Store{db: db}
}

func (s *Store) InsertTicket(t *session.Ticket) error {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(t); err != nil {
		return errors.Wrap(err, "failed to encode ticket")
	}

	err := s.db.Update(func(txn *badger.Txn) error {
		_, err := txn.Get([]byte(t.Token))
		if err == nil {
			return session.ErrCollidingToken
		}

		return txn.SetEntry(&badger.Entry{
			Key:       []byte(t.Token),
			Value:     buf.Bytes(),
			ExpiresAt: uint64(time.Now().Add(session.MaxAge).Unix()),
		})
	})

	if err != nil {
		return errors.Wrap(err, "unexpected badger error")
	}

	return nil
}

func (s *Store) FindTicket(token string) (*session.Ticket, error) {
	var ticket session.Ticket

	err := s.db.View(func(txn *badger.Txn) error {
		v, err := txn.Get([]byte(token))
		if err != nil {
			return err
		}

		return v.Value(func(buf []byte) error {
			return gob.NewDecoder(bytes.NewReader(buf)).Decode(&ticket)
		})
	})

	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			return nil, session.ErrTicketNotFound
		}

		return nil, errors.Wrap(err, "unexpected badger error")
	}

	return &ticket, nil
}

func (s *Store) InvalidateTicket(token string) {
	err := s.db.Update(func(txn *badger.Txn) error {
		err := txn.Delete([]byte(token))
		if err == nil || errors.Is(err, badger.ErrKeyNotFound) {
			return nil
		}
		return err
	})

	if err != nil {
		log.Panicln("BUG: badger failed to invalidate ticket:", err)
	}
}
