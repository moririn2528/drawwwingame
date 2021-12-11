package domain_test

import (
	"drawwwingame/domain"
	"drawwwingame/domain/valobj"
	"fmt"
	"strconv"
	"testing"
)

func getTestUsernGroup(t *testing.T, user_ver int) (*domain.User, *domain.GroupObject) {
	user, _ := getTestUser(t, user_ver)
	group_id, err := valobj.NewGroupIdInt(0)
	check(t, "NewGroupIdInt error", err, nil)
	group, err := user.UpdateGroup(group_id)
	check(t, "UpdateGroup error", err, nil)
	return user, group
}

func createUuids(t *testing.T, users ...*domain.User) []*valobj.UuidInt {
	s := []*valobj.UuidInt{}
	for i, u := range users {
		a, err := valobj.NewUuidInt(u.GetUuidInt())
		check(t, "createUuids: "+strconv.Itoa(i), err, nil)
		s = append(s, a)
	}
	return s
}

func uuidIn(t *testing.T, a []*valobj.UuidInt, b []*valobj.UuidInt) bool {
	m := make(map[int]bool)
	for _, u := range b {
		m[u.ToInt()] = true
	}
	for _, u := range a {
		_, ok := m[u.ToInt()]
		check(t, "uuidIn", ok, true)
		if !ok {
			return false
		}
	}
	return true

}

func TestGame(t *testing.T) {
	defer domain.SqlHandle.DeleteForTest()
	user0, _ := getTestUsernGroup(t, 0)
	user1, _ := getTestUsernGroup(t, 1)
	uuids := createUuids(t, user0, user1)
	group_id, err := valobj.NewGroupIdInt(0)
	check(t, "NewGroupIdInt error", err, nil)
	patterns := map[[3]int]error{
		// step, minute, writer_number
		{10, 10, 1}: nil,
		{10, 10, 2}: nil,
		{10, 10, 3}: domain.ErrorArg,
		{1, 1, 1}:   nil,
		{0, 10, 1}:  domain.ErrorArg,
		{10, 0, 1}:  domain.ErrorArg,
		{10, 10, 0}: domain.ErrorArg,
	}
	for arg, val := range patterns {
		game, err := domain.NewGame(group_id, arg[0], arg[1], arg[2])
		check(t, fmt.Sprintf("NewGame, %v", arg), err, val)
		if err != nil {
			continue
		}
		cnt, err := domain.SqlHandle.GetGroupUserCount(group_id)
		check(t, fmt.Sprintf("domain.SqlHandle.GetGroupUserCount error, %v", arg), err, nil)
		check(t, fmt.Sprintf("UserSize, %v", arg), game.UserSize(), cnt)
		check(t, fmt.Sprintf("writer number, %v", arg), len(game.GetWriterUuid()), arg[2])
		check(t, fmt.Sprintf("writer+answer number, %v", arg),
			len(game.GetWriterUuid())+len(game.GetAnswerUuid()), 2)
		uuidIn(t, game.GetAnswerUuid(), uuids)
		uuidIn(t, game.GetWriterUuid(), uuids)
		check(t, fmt.Sprintf("Start error, %v", arg), game.Start(), nil)
		check(t, fmt.Sprintf("End error, %v", arg), game.End(), nil)
		check(t, fmt.Sprintf("Finish error, %v", arg), game.Finish(), nil)
	}
}
