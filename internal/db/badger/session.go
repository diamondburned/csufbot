package badger

import (
	"bytes"
	"encoding/json"
	"log"
	"time"

	"github.com/dgraph-io/badger/v3"
	"github.com/diamondburned/csufbot/internal/csufbot/session"
	"github.com/pkg/errors"
)

type SessionStore struct {
	db *badger.DB
}

var _ session.Storer = (*SessionStore)(nil)

// NewSessionStore creates a new Badger store from an existing database.
func NewSessionStore(db *badger.DB) *SessionStore {
	return &SessionStore{db: db}
}

func sessionKey(typ session.TicketType, tok string) []byte {
	const prefix = "session"

	buf := bytes.Buffer{}
	buf.Grow(len(prefix) + 1 + len(tok))
	buf.WriteString(prefix)
	buf.WriteByte(byte(typ))
	buf.WriteString(tok)

	return buf.Bytes()
}

func (s *SessionStore) InsertTicket(t *session.Ticket) error {
	b, err := json.Marshal(t)
	if err != nil {
		return errors.Wrap(err, "failed ot encode ticket")
	}

	k := sessionKey(t.Type, t.Token)

	err = s.db.Update(func(txn *badger.Txn) error {
		_, err := txn.Get(k)
		if err == nil {
			return session.ErrCollidingToken
		}

		return txn.SetEntry(&badger.Entry{
			Key:       k,
			Value:     b,
			ExpiresAt: uint64(time.Now().Add(session.MaxAge).Unix()),
		})
	})

	if err != nil {
		return errors.Wrap(err, "unexpected badger error")
	}

	return nil
}

func (s *SessionStore) FindTicket(typ session.TicketType, tok string) (*session.Ticket, error) {
	var ticket session.Ticket

	k := sessionKey(typ, tok)

	err := s.db.View(func(txn *badger.Txn) error {
		v, err := txn.Get(k)
		if err != nil {
			return err
		}

		return v.Value(func(buf []byte) error {
			return json.Unmarshal(buf, &ticket)
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

func (s *SessionStore) InvalidateTicket(typ session.TicketType, tok string) {
	k := sessionKey(typ, tok)

	err := s.db.Update(func(txn *badger.Txn) error {
		err := txn.Delete(k)
		if err == nil || errors.Is(err, badger.ErrKeyNotFound) {
			return nil
		}
		return err
	})

	if err != nil {
		log.Panicln("BUG: badger failed to invalidate ticket:", err)
	}
}
