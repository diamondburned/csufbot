package session

import (
	"time"

	"github.com/diamondburned/arikawa/v2/discord"
	"github.com/pkg/errors"
)

// MaxAge is the maximum age of each ticket.
const MaxAge = 6 * time.Hour

// Storer is the generic interface to store session objects.
type Storer interface {
	// InsertTicket inserts a ticket. A ticket with a colliding token is
	// erroneous.
	InsertTicket(t *Ticket) error
	// FindTicket finds a ticket from the given token.
	FindTicket(token string) (*Ticket, error)
	// InvalidateTicket invalidates a ticket. If the database fails to
	// invalidate and cannot restore the database to sane state, then it is
	// allowed to panic.
	InvalidateTicket(token string)
}

var (
	// ErrTicketNotFound is returned if the ticket cannot be found from the Storer.
	ErrTicketNotFound = errors.New("ticket not found")
	// ErrCollidingToken is returned if the ticket's token overlaps another.
	ErrCollidingToken = errors.New("ticket has colliding token")
)

// Repository stores multiple authorization sessions. It is safe to be used
// concurrently.
type Repository struct {
	storer Storer
}

// NewRepository creates a new session repository.
func NewRepository(storer Storer) Repository {
	return Repository{
		storer: storer,
	}
}

// Register registers the given user ID into a newly registered ticket.
func (r Repository) Register(userID discord.UserID) (*Ticket, error) {
	var ticket = Ticket{UserID: userID}
	var err error

	// Try 10 times to randomize tokens.
	for i := 0; i < 10; i++ {
		ticket.Token, err = randToken()
		if err != nil {
			return nil, err
		}

		if err = r.storer.InsertTicket(&ticket); err != nil {
			// If we have a colliding token, then regenerate and retry.
			if errors.Is(err, ErrCollidingToken) {
				continue
			}

			return nil, errors.Wrap(err, "failed to save ticket")
		}

		break
	}

	return &ticket, err
}

// Ticket is a repository ticket.
type Ticket struct {
	Token  string
	UserID discord.UserID
}
