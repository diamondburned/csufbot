package badger

import (
	"github.com/dgraph-io/badger/v3"
	"github.com/diamondburned/arikawa/v2/discord"
	"github.com/diamondburned/csufbot/internal/csufbot"
	"github.com/diamondburned/csufbot/internal/lms"
	"github.com/pkg/errors"
)

type GuildStore struct {
	db *badger.DB
}

func (store *GuildStore) Guild(id discord.GuildID) (*csufbot.Guild, error) {
	var guild *csufbot.Guild
	return guild, unmarshal(store.db, "guild", u64b(uint64(id)), &guild)
}

func (store *GuildStore) SetCourses(g discord.GuildID, cs map[lms.CourseID]discord.RoleID) error {
	var guild *csufbot.Guild
	key := joinKeys("guild", u64b(uint64(g)))

	return store.db.Update(func(txn *badger.Txn) error {
		if err := unmarshalFromTxn(txn, key, &guild); err != nil {
			return err
		}

		guild.RoleMap = cs
		return marshalToTxn(txn, key, guild)
	})
}

func (store *GuildStore) GuildCourses(
	courseStore csufbot.CourseStorer, out map[discord.GuildID][]csufbot.Course) error {

	courseMap := make(map[lms.CourseID]csufbot.Course, len(out))

	err := store.db.View(func(txn *badger.Txn) error {
		for guildID := range out {
			var courses struct {
				Courses []lms.CourseID
			}

			key := joinKeys("guild", u64b(uint64(guildID)))

			if err := unmarshalFromTxn(txn, key, &courses); err != nil {
				return errors.Wrapf(err, "failed to get guild ID %d", guildID)
			}

			coursesList := make([]csufbot.Course, len(courses.Courses))
			out[guildID] = coursesList

			for i, id := range courses.Courses {
				courseMap[id] = csufbot.Course{}
				coursesList[i].ID = id
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	if err := courseStore.Courses(courseMap); err != nil {
		return errors.Wrap(err, "failed to get courses")
	}

	for _, guildCourses := range out {
		for i, course := range guildCourses {
			guildCourses[i] = courseMap[course.ID]
		}
	}

	return nil
}
