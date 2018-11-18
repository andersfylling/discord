package disgord

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/andersfylling/disgord/httd"
)

func TestUser_InterfaceImplementations(t *testing.T) {
	var u interface{} = &User{}

	//t.Run("UserInterface", func(t *testing.T) {
	//	if _, ok := u.(UserInterface); !ok {
	//		t.Error("User does not implement UserInterface")
	//	}
	//})

	t.Run("DeepCopier", func(t *testing.T) {
		if _, ok := u.(DeepCopier); !ok {
			t.Error("User does not implement DeepCopier")
		}
	})

	t.Run("Copier", func(t *testing.T) {
		if _, ok := u.(Copier); !ok {
			t.Error("User does not implement Copier")
		}
	})

	t.Run("DiscordSaver", func(t *testing.T) {
		if _, ok := u.(discordSaver); !ok {
			t.Error("User does not implement discordSaver")
		}
	})
}

func verifyUserMashaller(t *testing.T, file string) {
	data, err := ioutil.ReadFile(file)
	check(err, t)

	user := User{}
	err = httd.Unmarshal(data, &user)
	check(err, t)
}

func TestUserMarshalling(t *testing.T) {
	files := []string{
		"testdata/user/user1.json",
		"testdata/user/user2.json",
	}

	for _, file := range files {
		verifyUserMashaller(t, file)
	}
}

func TestUserPresence_InterfaceImplementations(t *testing.T) {
	var u interface{} = NewUserPresence()

	t.Run("Stringer", func(t *testing.T) {
		if _, ok := u.(fmt.Stringer); !ok {
			t.Error("UserPresence does not implement fmt.Stringer")
		}
	})

	t.Run("DeepCopier", func(t *testing.T) {
		if _, ok := u.(DeepCopier); !ok {
			t.Error("UserPresence does not implement DeepCopier")
		}
	})

	t.Run("Copier", func(t *testing.T) {
		if _, ok := u.(Copier); !ok {
			t.Error("UserPresence does not implement Copier")
		}
	})

	// TODO
	// t.Run("DiscordSaver", func(t *testing.T) {
	// 	if _, ok := u.(discordSaver); !ok {
	// 		t.Error("UserPresence does not implement discordSaver")
	// 	}
	// })
}

func equals(a, b *User) bool {
	ok := a.ID == b.ID && a.Username == b.Username && a.Discriminator == b.Discriminator && a.Email == b.Email && a.Token == b.Token && a.Verified == b.Verified && a.MFAEnabled == b.MFAEnabled && a.Bot == b.Bot
	if a.Avatar == nil && b.Avatar != a.Avatar {
		ok = false
	}
	if b.Avatar == nil && b.Avatar != a.Avatar {
		ok = false
	}
	if a.Avatar != nil && b.Avatar != nil && *a.Avatar != *b.Avatar {
		ok = false
	}

	return ok
}

type userCopyOverToCacheTestSets struct {
	user  User
	bytes []byte
}

func TestUser_copyOverToCache(t *testing.T) {
	datas := make([]userCopyOverToCacheTestSets, 6)

	user1 := User{}
	unmarshal([]byte(`{"id": "133218433276182528","username":"alak","discriminator":"1149","avatar":"38d04eba240fa3cad581a947025644ad","verified":true}`), &user1)
	datas[0] = userCopyOverToCacheTestSets{user1, []byte(`{"id": "133218433276182528","username":"alak","discriminator":"1149","avatar":"38d04eba240fa3cad581a947025644ad","verified":true}`)}

	user2 := User(user1)
	user2.Bot = true
	datas[1] = userCopyOverToCacheTestSets{user2, []byte(`{"id": "133218433276182528","username":"alak","discriminator":"1149","avatar":"38d04eba240fa3cad581a947025644ad","bot":true}`)}

	user3 := User(user2)
	user3.Discriminator = Discriminator(1849)
	datas[2] = userCopyOverToCacheTestSets{user3, []byte(`{"id": "133218433276182528","username":"alak","discriminator":"1849","avatar":"38d04eba240fa3cad581a947025644ad"}`)}

	user4 := User(user3)
	user4.Avatar = nil
	datas[3] = userCopyOverToCacheTestSets{user4, []byte(`{"id": "133218433276182528","username":"alak","discriminator":"1849","avatar":null}`)}

	user5 := User(user4)
	datas[4] = userCopyOverToCacheTestSets{user5, []byte(`{"id": "133218433276182528"}`)}

	user6 := User(user5)
	user6.Username = "sdfsd"
	a := "aaaaaaaaa"
	user6.Avatar = &a
	user6.Discriminator = Discriminator(1249)
	user6.Verified = false
	datas[5] = userCopyOverToCacheTestSets{user6, []byte(`{"id": "133218433276182528","username":"sdfsd","discriminator":"1249","avatar":"aaaaaaaaa","verified":false}`)}

	var cache User
	var err error
	for i := range datas {
		bytes := datas[i].bytes
		expected := datas[i].user
		var user User
		err = unmarshal(bytes, &user)
		if err != nil {
			t.Error(err)
			return
		}

		user.copyOverToCache(&cache)

		//fmt.Printf("##: %+v\n", cacheLink)

		if !equals(&cache, &expected) {
			t.Errorf("different users. \nGot \t%+v, \nWants \t%+v", cache, expected)
		}
	}
}

func TestGetCurrentUser(t *testing.T) {
	client, _, err := createTestRequester()
	if err != nil {
		t.Skip()
		return
	}

	_, err = GetCurrentUser(client)
	if err != nil {
		t.Error(err)
	}
}
func TestGetUser(t *testing.T) {
	client, _, err := createTestRequester()
	if err != nil {
		t.Skip()
		return
	}

	userID := NewSnowflake(140413331470024704)
	user, err := GetUser(client, userID)
	if err != nil {
		t.Error(err)
	}

	if user.ID != userID {
		t.Error("user ID missmatch")
	}
}
func TestModifyCurrentUser(t *testing.T) {
	client, _, err := createTestRequester()
	if err != nil {
		t.Skip()
		return
	}

	// this has been verified to work
	// however, you cannot change username often so this is
	// can give an error

	var originalUsername string
	t.Run("getting original username", func(t *testing.T) {
		user, err := GetCurrentUser(client)
		if err != nil {
			t.Error(err)
		}

		originalUsername = user.Username
	})

	t.Run("changing username", func(t *testing.T) {
		if originalUsername == "" {
			t.Skip()
			return
		}
		params := &ModifyCurrentUserParams{}
		params.SetUsername("sldfhksghs")
		_, err := ModifyCurrentUser(client, params)
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("resetting username", func(t *testing.T) {
		if originalUsername == "" {
			t.Skip()
			return
		}
		params := &ModifyCurrentUserParams{}
		params.SetUsername(originalUsername)
		_, err := ModifyCurrentUser(client, params)
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("changing avatar", func(t *testing.T) {
		params := &ModifyCurrentUserNick{}
		file, _ := os.Open("./gopher.png")
		defer file.Close()
		params.SetAvatarImage(file)
		_, err := ModifyCurrentUser(client, params)
		if err != nil {
			t.Error(err)
		}
	})
}
func TestGetCurrentUserGuildsParams(t *testing.T) {
	params := &GetCurrentUserGuildsParams{}
	var wants string

	wants = ""
	verifyQueryString(t, params, wants)

	s := "438543957"
	params.Before, _ = GetSnowflake(s)
	wants = "?before=" + s
	verifyQueryString(t, params, wants)

	params.Limit = 6
	wants += "&limit=6"
	verifyQueryString(t, params, wants)

	params.Limit = 0
	wants = "?before=" + s
	verifyQueryString(t, params, wants)
}
func TestLeaveGuild(t *testing.T) {
	// Nope. Not gonna automate this.
}
func TestUserDMs(t *testing.T) {
	// TODO
}

func TestCreateDM(t *testing.T) {
	client, _, err := createTestRequester()
	if err != nil {
		t.Skip()
		return
	}

	channel, err := CreateDM(client, 228846961774559232)
	if err != nil {
		t.Error(err)
	}

	if channel == nil {
		t.Error("channel was nil")
	}
}
func TestCreateGroupDM(t *testing.T) {
	// TODO
}
func TestGetUserConnections(t *testing.T) {
	// Missing OAuth2
}
