package infra

import (
	"database/sql"
	"drawwwingame/domain"
	"drawwwingame/domain/valobj"
	"log"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type SqlHandler struct {
	db *sqlx.DB
}

func StringsToTableString(table string, strs []string) string {
	res := make([]string, len(strs))
	for i, key := range strs {
		res[i] = table + "." + key
	}
	return strings.Join(res, ",")
}

func NewSqlHandler() (*SqlHandler, error) {
	handle := new(SqlHandler)
	var err error
	handle.db, err = sqlx.Open(domain.GetDatabaseInfo())
	if err != nil {
		domain.Log(err)
		return nil, err
	}
	return handle, nil
}

func (handle *SqlHandler) insertSqlMany(table string, args []map[string]interface{}) (sql.Result, error) {
	argname := []string{}
	inserts := []string{}
	namedarg := make(map[string]interface{})
	for i, arg := range args {
		str := []string{}
		for k, val := range arg {
			if i == 0 {
				argname = append(argname, k)
			}
			key := k + strconv.Itoa(i)
			str = append(str, ":"+key)
			namedarg[key] = val
		}
		inserts = append(inserts, "("+strings.Join(str, ",")+")")
	}
	exec_str := "INSERT INTO " + table + "(" + strings.Join(argname, ",") + ") VALUES " + strings.Join(inserts, ",")
	return handle.db.NamedExec(exec_str, namedarg)
}

func (handle *SqlHandler) insertSql(table string, arg map[string]interface{}) (sql.Result, error) {
	return handle.insertSqlMany(table, []map[string]interface{}{arg})
}

func (handle *SqlHandler) updateSql(table string, arg map[string]interface{}, search_arg []string) (sql.Result, error) {
	arg_str := []string{}
	search_arg_str := []string{}
	for k := range arg {
		arg_str = append(arg_str, k+"=:"+k)
	}
	for _, s := range search_arg {
		search_arg_str = append(search_arg_str, s+"=:"+s)
	}
	exec_str := "UPDATE " + table + " SET " + strings.Join(arg_str, ",")
	if len(search_arg_str) > 0 {
		exec_str += " WHERE " + strings.Join(search_arg_str, " AND ")
	}
	return handle.db.NamedExec(exec_str, arg)
}

func checkAffectRow(res sql.Result, cnt int) error {
	affect, err := res.RowsAffected()
	if err != nil {
		domain.Log(err)
		return domain.ErrorInternal
	}
	if int(affect) != cnt {
		domain.LogStringf("affect row error, %v, %v", affect, cnt)
		return domain.ErrorInternal
	}
	return nil
}

func (handle *SqlHandler) Init() error {
	var info []struct {
		Name  string `db:"name"`
		Value string `db:"value"`
	}
	err := handle.db.Select(&info, "SELECT name, value FROM secret")
	if err != nil {
		domain.Log(err)
		return domain.ErrorInternal
	}
	m := make(map[string]string)
	for _, v := range info {
		m[v.Name] = v.Value
	}
	domain.SetInternalVar(m)

	//set message id
	var cnt int
	err = handle.db.Get(&cnt, "SELECT IFNULL(MAX(id),-1) FROM message")
	if err != nil {
		domain.Log(err)
		return domain.ErrorInternal
	}
	valobj.SetMessageId(cnt + 1)
	return err
}

func (handle *SqlHandler) GetUserById(input_uuid *valobj.UuidInt, input_tempid *valobj.TempIdString) (*domain.SqlUser, error) {
	user := &domain.SqlUser{}
	err := handle.db.Get(user, "SELECT "+strings.Join(domain.SqlUserMapKey, ",")+" FROM user WHERE uuid = ? AND tempid = ?",
		input_uuid.ToInt(), input_tempid.ToString(),
	)
	if err != nil {
		domain.Log(err)
		return nil, err
	}
	return user, nil
}

func (handle *SqlHandler) GetUsersByUuid(uuids []*valobj.UuidInt) ([]*domain.User, error) {
	var err error
	users := []domain.SqlUser{}
	querys := make([]string, len(uuids))
	args := make([]interface{}, len(uuids))
	for i, u := range uuids {
		querys[i] = "SELECT " + strings.Join(domain.SqlUserMapKey, ",") + " FROM user WHERE uuid = ?"
		args[i] = u.ToInt()
	}
	err = handle.db.Select(&users,
		strings.Join(querys, " UNION "), args...,
	)
	if err != nil {
		domain.Log(err)
		return nil, err
	}
	res := make([]*domain.User, len(users))
	for i, u := range users {
		res[i], err = u.ToUser()
		if err != nil {
			domain.Log(err)
			return nil, err
		}
	}
	return res, nil
}

func (handle *SqlHandler) GetUser(input_name *valobj.NameString, password *valobj.PasswordString) (*domain.SqlUser, error) {
	user := &domain.SqlUser{}
	err := handle.db.Get(user, "SELECT "+strings.Join(domain.SqlUserMapKey, ",")+" FROM user WHERE name=? AND password=?",
		input_name.ToString(), password.ToString(),
	)
	if err != nil {
		domain.Log(err)
		return nil, err
	}
	return user, nil
}

func (handle *SqlHandler) CreateUser(user *domain.User, password *valobj.PasswordString) error {
	user_map := user.GetAllMap()
	user_map["password"] = password.ToString()
	rows, err := handle.db.Queryx("SELECT COUNT(*) FROM user WHERE name=? UNION ALL "+
		"SELECT COUNT(*) FROM user WHERE email=? UNION ALL "+
		"SELECT COUNT(*) FROM user WHERE uuid=?",
		user_map["name"], password.ToString(), user_map["email"],
	)
	if err != nil {
		domain.Log(err)
		return err
	}
	defer rows.Close()
	i := 0
	for rows.Next() {
		var num int
		err := rows.Scan(&num)
		if err != nil {
			domain.Log(err)
			return err
		}
		if num > 0 {
			switch i {
			case 0:
				return domain.ErrorDuplicateUserName
			case 1:
				return domain.ErrorDuplicateUserEmail
			case 2:
				return domain.ErrorDuplicateUserId
			}
		}
		i++
	}
	err = rows.Err()
	if err != nil {
		domain.Log(err)
		return domain.ErrorInternal
	}
	if i != 3 {
		domain.LogStringf("rows over 3")
		return domain.ErrorInternal
	}
	res, err := handle.insertSql("user", user_map)
	if err != nil {
		domain.Log(err)
		return domain.ErrorInternal
	}
	row_affect, err := res.RowsAffected()
	if err != nil {
		domain.Log(err)
		return domain.ErrorInternal
	}
	if row_affect != 1 {
		domain.LogStringf("no inserted")
		return domain.ErrorInternal
	}

	return nil
}

func (handle *SqlHandler) UpdateUser(user *domain.User) error {
	res, err := handle.updateSql("user", user.GetAllMap(), []string{"uuid"})
	if err != nil {
		domain.Log(err)
		return err
	}
	row_affect, err := res.RowsAffected()
	if err != nil {
		domain.Log(err)
		return err
	}
	if row_affect != 1 {
		domain.LogStringf("no updated")
		return domain.ErrorInternal
	}
	return nil
}

func (handle *SqlHandler) DeleteForTest() {
	deleteFromTable := func(email string) error {
		var uuid int
		var cnt int
		err := handle.db.Get(&cnt, "SELECT COUNT(uuid) FROM user WHERE email=?", email)
		if err != nil {
			domain.Log(err)
			return err
		}
		if cnt == 0 {
			return nil
		}
		err = handle.db.Get(&uuid, "SELECT uuid FROM user WHERE email=?", email)
		if err != nil {
			domain.Log(err)
			return err
		}
		tables := []string{"user", "message", "message_mark", "user_group"}
		for _, t := range tables {
			_, err = handle.db.Exec("DELETE FROM "+t+" WHERE uuid=?", uuid)
			if err != nil {
				log.Println("table: " + t + ", uuid: " + strconv.Itoa(uuid))
				domain.Log(err)
				return err
			}
		}
		return nil
	}
	deleteFromTable(domain.GetTestGmailAddress(0))
	deleteFromTable(domain.GetTestGmailAddress(1))
}

func (handle *SqlHandler) CreateMessage(msg *domain.MessageObject) error {
	res, err := handle.insertSql("message", msg.GetMap())
	if err != nil {
		domain.Log(err)
		return err
	}
	row_affect, err := res.RowsAffected()
	if err != nil {
		domain.Log(err)
		return err
	}
	if row_affect != 1 {
		domain.LogStringf("no inserted")
		return domain.ErrorInternal
	}
	return nil
}

func (handle *SqlHandler) GetMessage(id *valobj.MessageId) (*domain.MessageObject, error) {
	sql_msg := domain.SqlMessage{}
	err := handle.db.Get(&sql_msg, "SELECT "+strings.Join(domain.MessageObjectKey, ",")+" FROM message WHERE id=?",
		id.ToInt())
	if err != nil {
		domain.Log(err)
		return nil, domain.ErrorInternal
	}
	msg, err := sql_msg.ToMessageObject()
	if err != nil {
		domain.Log(err)
		return nil, domain.ErrorInternal
	}
	return msg, nil
}

func (handle *SqlHandler) GetGameMessage(id *valobj.GroupIdInt, tim time.Time) ([]*domain.MessageObject, []*domain.MessageMark, error) {
	sql_msg := []domain.SqlMessage{}
	err := handle.db.Select(&sql_msg, "SELECT "+strings.Join(domain.MessageObjectKey, ",")+
		" FROM message WHERE created_at>=? AND group_id=?",
		tim, id.ToInt(),
	)
	if err != nil {
		domain.Log(err)
		return nil, nil, err
	}
	sql_mark := []domain.SqlMessageMark{}
	err = handle.db.Select(&sql_mark, "SELECT "+strings.Join(domain.MessageMarkKey, ",")+
		" FROM message_mark WHERE created_at>=? AND group_id=?",
		tim, id.ToInt(),
	)
	if err != nil {
		domain.Log(err)
		return nil, nil, err
	}
	msg := make([]*domain.MessageObject, len(sql_msg))
	mark := make([]*domain.MessageMark, len(sql_mark))
	for i, m := range sql_msg {
		msg[i], err = m.ToMessageObject()
		if err != nil {
			domain.Log(err)
			return nil, nil, err
		}
	}
	for i, m := range sql_mark {
		mark[i], err = m.ToMessageMark()
		if err != nil {
			domain.Log(err)
			return nil, nil, err
		}
	}
	return msg, mark, nil
}

func (handle *SqlHandler) GetGroupUserCount(group_id *valobj.GroupIdInt) (int, error) {
	var cnt int
	err := handle.db.QueryRowx("SELECT COUNT(uuid) FROM user_group WHERE group_id=?", group_id.ToInt()).Scan(&cnt)
	if err != nil {
		domain.Log(err)
		return -1, domain.ErrorInternal
	}
	return cnt, nil
}

func (handle *SqlHandler) GetUserInGroup(group *valobj.GroupIdInt) ([]*domain.User, error) {
	users := []*domain.SqlUser{}
	err := handle.db.Select(&users, "SELECT "+StringsToTableString("user", domain.SqlUserMapKey)+" FROM user_group WHERE group_id=? "+
		"LEFT OUTER JOIN user ON user_group.uuid=user.uuid", group.ToInt())
	res := []*domain.User{}
	if err != nil {
		domain.Log(err)
		return res, domain.ErrorInternal
	}
	for _, user := range users {
		u, err := user.ToUser()
		if err != nil {
			domain.Log(err)
			return []*domain.User{}, domain.ErrorInternal
		}
		res = append(res, u)
	}
	return res, nil
}

func (handle *SqlHandler) GetUuidInGroup(group *valobj.GroupIdInt) ([]*valobj.UuidInt, error) {
	uuids := []int{}
	err := handle.db.Select(&uuids, "SELECT uuid FROM user_group WHERE group_id=?", group.ToInt())
	res := []*valobj.UuidInt{}
	if err != nil {
		domain.Log(err)
		return res, domain.ErrorInternal
	}
	for _, u := range uuids {
		uuid, err := valobj.NewUuidInt(u)
		if err != nil {
			domain.Log(err)
			return []*valobj.UuidInt{}, domain.ErrorInternal
		}
		res = append(res, uuid)
	}
	return res, nil
}

func mapKeyToString(m map[string]interface{}) string {
	s := []string{}
	for key := range m {
		s = append(s, key)
	}
	return strings.Join(s, ",")
}

func (handle *SqlHandler) SetMessageMark(mark *domain.MessageMark) error {
	var cnt int
	m := mark.GetMap()
	condition := "uuid=" + strconv.Itoa(m["uuid"].(int)) + " AND message_id=" + strconv.Itoa(m["message_id"].(int))
	err := handle.db.Get(&cnt, "SELECT COUNT(uuid) FROM message_mark WHERE "+condition)
	if err != nil {
		domain.Log(err)
		return domain.ErrorInternal
	}
	if cnt > 1 {
		domain.LogStringf("Marks too much in sql")
		return domain.ErrorInternal
	}
	if cnt == 1 {
		_, err = handle.updateSql("message_mark", m, []string{"uuid", "message_id"})
		if err != nil {
			domain.Log(err)
			return domain.ErrorInternal
		}
		return nil
	}
	_, err = handle.insertSql("message_mark", m)
	if err != nil {
		domain.Log(err)
		return domain.ErrorInternal
	}
	return nil
}

func (handle *SqlHandler) GetMarksOnMessage(id *valobj.MessageId) ([]int, error) {
	marks := []string{}
	s := make([]int, len(valobj.MessageMarksAllString))
	err := handle.db.Select(&marks, "SELECT mark FROM message_mark WHERE message_id=?", id.ToInt())
	if err != nil {
		domain.Log(err)
		return s, domain.ErrorInternal
	}
	for _, m := range marks {
		idx := strings.Index(valobj.MessageMarksAllString, m)
		if idx < 0 || len(valobj.MessageMarksAllString) <= idx {
			domain.LogStringf("error index: %v", idx)
			return s, domain.ErrorInternal
		}
		s[idx]++
	}
	return s, nil
}

func (handle *SqlHandler) GetGroupObjectByUuid(uuid *valobj.UuidInt) (*domain.GroupObject, error) {
	group := []domain.SqlGroupUser{}
	err := handle.db.Select(&group, "SELECT "+strings.Join(domain.GroupObjectMapKeys, ",")+" FROM user_group WHERE uuid="+strconv.Itoa(uuid.ToInt()))
	if err != nil {
		domain.Log(err)
		return nil, domain.ErrorInternal
	}
	if len(group) != 1 {
		domain.LogStringf("group size error: %v", len(group))
		return nil, domain.ErrorInternal
	}
	return group[0].ToGroupObject()
}
func (handle *SqlHandler) GetGroupObjectsByUuid(uuids []*valobj.UuidInt) ([]*domain.GroupObject, error) {
	groups := []domain.SqlGroupUser{}
	queries := make([]string, len(uuids))
	args := make([]interface{}, len(uuids))
	for i, u := range uuids {
		queries[i] = "SELECT " + strings.Join(domain.GroupObjectMapKeys, ",") + " FROM user_group WHERE uuid=?"
		args[i] = u.ToInt()
	}
	err := handle.db.Select(&groups, strings.Join(queries, " UNION "), args...)
	if err != nil {
		domain.Log(err)
		return nil, domain.ErrorInternal
	}
	res := make([]*domain.GroupObject, len(groups))
	for i, g := range groups {
		res[i], err = g.ToGroupObject()
		if err != nil {
			domain.Log(err)
			return nil, domain.ErrorInternal
		}
	}
	return res, nil
}

func (handle *SqlHandler) GetnSetGroupObjectById(uuid *valobj.UuidInt, group_id *valobj.GroupIdInt) (*domain.GroupObject, error) {
	group := []domain.SqlGroupUser{}
	err := handle.db.Select(&group, "SELECT "+strings.Join(domain.GroupObjectMapKeys, ",")+" FROM user_group WHERE uuid="+strconv.Itoa(uuid.ToInt()))
	if err != nil {
		domain.Log(err)
		return nil, domain.ErrorInternal
	}
	if len(group) > 1 {
		domain.LogStringf("group size error: %v", len(group))
		return nil, domain.ErrorInternal
	}
	new_group := domain.NewGroupObject(uuid, group_id, valobj.NewGroupRoleNoSet())
	if len(group) == 1 {
		g, err := group[0].ToGroupObject()
		if err != nil {
			domain.Log(err)
			return nil, domain.ErrorInternal
		}
		if g.GetGroupId().Equal(group_id) {
			return g, nil
		}
		res, err := handle.updateSql("user_group", new_group.GetMap(), []string{"uuid"})
		if err != nil {
			domain.Log(err)
			return nil, domain.ErrorInternal
		}
		if err = checkAffectRow(res, 1); err != nil {
			return nil, err
		}
		return new_group, nil
	}
	res, err := handle.insertSql("user_group", new_group.GetMap())
	if err != nil {
		domain.Log(err)
		return nil, domain.ErrorInternal
	}
	if err = checkAffectRow(res, 1); err != nil {
		return nil, err
	}
	return new_group, nil
}

func (handle *SqlHandler) UpdateGroupObject(group *domain.GroupObject) error {
	var cnt int
	err := handle.db.Get(&cnt, "SELECT COUNT(uuid) FROM user_group WHERE uuid="+strconv.Itoa(group.GetUuid().ToInt()))
	if err != nil {
		domain.Log(err)
		return domain.ErrorInternal
	}
	if cnt > 1 {
		domain.LogStringf("group size error: %v", cnt)
		return domain.ErrorInternal
	}
	var res sql.Result
	if cnt == 1 {
		res, err = handle.updateSql("user_group", group.GetMap(), []string{"uuid"})
	} else {
		res, err = handle.insertSql("user_group", group.GetMap())
	}
	if err != nil {
		domain.Log(err)
		return err
	}
	if err = checkAffectRow(res, 1); err != nil {
		return err
	}
	return nil
}

func (handle *SqlHandler) GetAllGroupObjectByGroupId(group_id *valobj.GroupIdInt) ([]*domain.GroupObject, error) {
	s := []domain.SqlGroupUser{}
	group := []*domain.GroupObject{}
	err := handle.db.Select(&s, "SELECT "+strings.Join(domain.GroupObjectMapKeys, ",")+" FROM user_group WHERE group_id=?", group_id.ToInt())
	if err != nil {
		domain.Log(err)
		return group, domain.ErrorInternal
	}
	for _, g := range s {
		gobj, err := g.ToGroupObject()
		if err != nil {
			domain.Log(err)
			return group, domain.ErrorInternal
		}
		group = append(group, gobj)
	}
	return group, nil
}

func (handle *SqlHandler) SetAllGroupObject(group []*domain.GroupObject) error {
	uuid_list := []string{}
	for _, g := range group {
		uuid_list = append(uuid_list, strconv.Itoa(g.GetGroupId().ToInt()))
	}
	res, err := handle.db.NamedExec("DELETE FROM user_group WHERE uuid IN (:uuids)", map[string]interface{}{
		"uuids": strings.Join(uuid_list, ","),
	})
	if err != nil {
		domain.Log(err)
		return domain.ErrorInternal
	}
	if err := checkAffectRow(res, len(uuid_list)); err != nil {
		domain.Log(err)
		return err
	}
	var allmaps []map[string]interface{}
	for _, g := range group {
		allmaps = append(allmaps, g.GetMap())
	}
	res, err = handle.insertSqlMany("user_group", allmaps)
	if err != nil {
		domain.Log(err)
		return domain.ErrorInternal
	}
	if err := checkAffectRow(res, len(uuid_list)); err != nil {
		domain.Log(err)
		return err
	}
	return nil
}

func (handle *SqlHandler) GetUserFromMessageId(id *valobj.MessageId) (*domain.User, error) {
	var user domain.SqlUser
	err := handle.db.Get(&user, "SELECT "+StringsToTableString("user", domain.SqlUserMapKey)+
		" FROM message LEFT JOIN user ON message.id=? AND message.uuid=user.uuid WHERE message.id=?",
		id.ToInt(), id.ToInt())
	if err != nil {
		domain.Log(err)
		return nil, domain.ErrorInternal
	}
	return user.ToUser()
}
