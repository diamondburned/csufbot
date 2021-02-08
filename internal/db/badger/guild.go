package badger

import (
	"github.com/dgraph-io/badger/v3"
	"github.com/diamondburned/arikawa/v2/discord"
	"github.com/diamondburned/csufbot/csufbot"
	"github.com/diamondburned/csufbot/csufbot/lms"
	"github.com/pkg/errors"
)

type GuildStore struct {
	*needDatabase
}

func (store GuildStore) Guild(id discord.GuildID) (*csufbot.Guild, error) {
	var guild *csufbot.Guild
	return guild, store.unmarshal("guild", u64b(uint64(id)), &guild)
}

func (store GuildStore) SetCourses(g discord.GuildID, cs map[lms.CourseID]discord.RoleID) error {
	var guild *csufbot.Guild
	key := joinKeys("guild", u64b(uint64(g)))

	return store.db.Update(func(txn *badger.Txn) error {
		if err := unmarshalFromTxn(txn, key, &guild); err != nil {
			if !errors.Is(err, badger.ErrKeyNotFound) {
				return errors.Wrap(err, "failed to get previous guild state")
			}

			guild = &csufbot.Guild{ID: g}
		}

		guild.RoleMap = cs
		return marshalToTxn(txn, key, guild)
	})
}
