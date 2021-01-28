// Package badger provides the BadgerDB backend for storing data.
package badger

import (
	"github.com/dgraph-io/badger/v3"
	"github.com/dgraph-io/badger/v3/options"
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
