// Package badger provides the BadgerDB backend for storing data.
package badger

import (
	"bytes"
	"encoding/binary"
	"encoding/json"

	"github.com/dgraph-io/badger/v3"
	"github.com/dgraph-io/badger/v3/options"
	"github.com/diamondburned/csufbot/csufbot"
)

// Open opens a new Badger database.
func Open(path string) (*badger.DB, error) {
	opts := badger.LSMOnlyOptions(path)
	opts = opts.WithLoggingLevel(badger.WARNING)
	opts.ChecksumVerificationMode = options.OnTableRead
	opts.CompactL0OnClose = true
	opts.Compression = options.Snappy
	opts.DetectConflicts = false

	return badger.Open(opts)
}

// Wrap wraps an existing database.
func Wrap(db *badger.DB) csufbot.Store {
	return csufbot.Store{
		Courses: &CourseStore{db},
		Guilds:  &GuildStore{db},
		Users:   &UserStore{db},
	}
}

// New creates a new Badger database and wraps it into a Store.
func New(path string) (csufbot.Store, error) {
	db, err := Open(path)
	if err != nil {
		return csufbot.Store{}, nil
	}

	return Wrap(db), nil
}

var nilBytes = []byte{0}

func joinKeys(prefix string, key []byte) []byte {
	keyBuf := bytes.Buffer{}
	keyBuf.Grow(len(prefix) + 1 + len(key))
	keyBuf.WriteString(prefix)
	keyBuf.WriteByte(0)
	keyBuf.Write(key)

	return keyBuf.Bytes()
}

func unmarshalString(db *badger.DB, prefix, key string, v interface{}) error {
	return unmarshal(db, prefix, []byte(key), v)
}

func unmarshal(db *badger.DB, prefix string, key []byte, v interface{}) error {
	keyBuf := joinKeys(prefix, key)
	return db.View(func(txn *badger.Txn) error {
		return unmarshalFromTxn(txn, keyBuf, v)
	})
}

func unmarshalFromTxn(txn *badger.Txn, k []byte, v interface{}) error {
	t, err := txn.Get(k)
	if err != nil {
		return err
	}
	return t.Value(func(value []byte) error {
		return json.Unmarshal(value, v)
	})
}

func marshalToTxn(txn *badger.Txn, k []byte, v interface{}) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}

	return txn.Set(k, b)
}

var intEndianness = binary.LittleEndian

func u64b(u64 uint64) []byte {
	buf := make([]byte, 8)
	intEndianness.PutUint64(buf, u64)
	return buf
}
