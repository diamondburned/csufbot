package sync

import (
	"testing"

	"github.com/diamondburned/arikawa/v2/discord"
	"github.com/diamondburned/csufbot/csufbot"
	"github.com/diamondburned/csufbot/csufbot/lms"
	"github.com/diamondburned/csufbot/internal/web"
	"github.com/hexops/autogold"
)

func guildsInServiceFakeData() ([]discord.Guild, web.RenderConfig) {
	guilds := []discord.Guild{
		{ID: 1, Name: "joe"},
		{ID: 2, Name: "mom"},
		{ID: 3, Name: "bruh"},
		{ID: 4, Name: "foo"},
		{ID: 5, Name: "bar"},
		{ID: 6, Name: "i'm running out"},
		{ID: 7, Name: "of ideas!!!"},
	}

	lmsServices := []web.LMSService{
		web.LMSService{Service: hoster{host: "abc.com"}},
		web.LMSService{Service: hoster{host: "bbc.com"}},
		web.LMSService{Service: hoster{host: "def.com"}},
	}

	mockGuildCourses := mockGuildCourses{
		courses: map[discord.GuildID][]csufbot.Course{
			1: {
				{Course: lms.Course{ID: "1", Name: "how to code"}, ServiceHost: "abc.com"},
				{Course: lms.Course{ID: "2", Name: "how to delete"}, ServiceHost: "bbc.com"},
				{Course: lms.Course{ID: "3", Name: "how to undo"}, ServiceHost: "def.com"},
				{Course: lms.Course{ID: "4", Name: "how to paste"}, ServiceHost: "abc.com"},
			},
			2: {
				{Course: lms.Course{ID: "2", Name: "how to delete"}, ServiceHost: "bbc.com"},
				{Course: lms.Course{ID: "3", Name: "how to undo"}, ServiceHost: "def.com"},
			},
			4: {
				{Course: lms.Course{ID: "1", Name: "how to code"}, ServiceHost: "abc.com"},
				{Course: lms.Course{ID: "3", Name: "how to undo"}, ServiceHost: "def.com"},
				{Course: lms.Course{ID: "4", Name: "how to paste"}, ServiceHost: "abc.com"},
			},
			5: {
				{Course: lms.Course{ID: "2", Name: "how to delete"}, ServiceHost: "bbc.com"},
				{Course: lms.Course{ID: "3", Name: "how to undo"}, ServiceHost: "def.com"},
				{Course: lms.Course{ID: "4", Name: "how to paste"}, ServiceHost: "abc.com"},
				{Course: lms.Course{ID: "5", Name: "how to bruh"}, ServiceHost: "abc.com"},
			},
			7: {
				{Course: lms.Course{ID: "6", Name: "how to what"}, ServiceHost: "bbc.com"},
			},
		},
	}

	return guilds, web.RenderConfig{
		Services: lmsServices,
		Store: csufbot.Store{
			Guilds: mockGuildCourses,
		},
	}
}

func BenchmarkGuildsInServices(b *testing.B) {
	guilds, config := guildsInServiceFakeData()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		guildsInServices(guilds, config)
	}
}

func TestGuildsInServices(t *testing.T) {
	guilds, config := guildsInServiceFakeData()

	want := autogold.Want("", map[lms.Host][]guildCourses{
		"abc.com": {
			{
				Guild: &guilds[5-1],
				Courses: []csufbot.Course{
					{Course: lms.Course{ID: "4", Name: "how to paste"}, ServiceHost: "abc.com"},
					{Course: lms.Course{ID: "5", Name: "how to bruh"}, ServiceHost: "abc.com"},
				},
			},
			{
				Guild: &guilds[4-1],
				Courses: []csufbot.Course{
					{Course: lms.Course{ID: "1", Name: "how to code"}, ServiceHost: "abc.com"},
					{Course: lms.Course{ID: "4", Name: "how to paste"}, ServiceHost: "abc.com"},
				},
			},
			{
				Guild: &guilds[1-1],
				Courses: []csufbot.Course{
					{Course: lms.Course{ID: "1", Name: "how to code"}, ServiceHost: "abc.com"},
					{Course: lms.Course{ID: "4", Name: "how to paste"}, ServiceHost: "abc.com"},
				},
			},
		},
		"bbc.com": {
			{
				Guild: &guilds[5-1],
				Courses: []csufbot.Course{
					{Course: lms.Course{ID: "2", Name: "how to delete"}, ServiceHost: "bbc.com"},
				},
			},
			{
				Guild: &guilds[1-1],
				Courses: []csufbot.Course{
					{Course: lms.Course{ID: "2", Name: "how to delete"}, ServiceHost: "bbc.com"},
				},
			},
			{
				Guild: &guilds[2-1],
				Courses: []csufbot.Course{
					{Course: lms.Course{ID: "2", Name: "how to delete"}, ServiceHost: "bbc.com"},
				},
			},
			{
				Guild: &guilds[7-1],
				Courses: []csufbot.Course{
					{Course: lms.Course{ID: "6", Name: "how to what"}, ServiceHost: "bbc.com"},
				},
			},
		},
		"def.com": {
			{
				Guild: &guilds[5-1],
				Courses: []csufbot.Course{
					{Course: lms.Course{ID: "3", Name: "how to undo"}, ServiceHost: "def.com"},
				},
			},
			{
				Guild: &guilds[4-1],
				Courses: []csufbot.Course{
					{Course: lms.Course{ID: "3", Name: "how to undo"}, ServiceHost: "def.com"},
				},
			},
			{
				Guild: &guilds[1-1],
				Courses: []csufbot.Course{
					{Course: lms.Course{ID: "3", Name: "how to undo"}, ServiceHost: "def.com"},
				},
			},
			{
				Guild: &guilds[2-1],
				Courses: []csufbot.Course{
					{Course: lms.Course{ID: "3", Name: "how to undo"}, ServiceHost: "def.com"},
				},
			},
		},
	})

	want.Equal(t, guildsInServices(guilds, config))
}

type mockGuildCourses struct {
	csufbot.GuildStorer
	courses map[discord.GuildID][]csufbot.Course
}

func (gcs mockGuildCourses) GuildCourses(c map[discord.GuildID][]csufbot.Course) error {
	for id, courses := range gcs.courses {
		c[id] = courses
	}
	return nil
}

type hoster struct {
	// whatever, nil is fine
	lms.Service
	host lms.Host
}

func (h hoster) Host() lms.Host { return h.host }
