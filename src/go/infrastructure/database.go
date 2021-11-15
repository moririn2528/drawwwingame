package infra

import (
	"drawwwingame/domain"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type SqlHandler struct {
	db *sqlx.DB
}

func NewSqlHandler() (*SqlHandler, error) {
	handle := new(SqlHandler)
	var err error
	handle.db, err = sqlx.Open(domain.DATABASE, domain.DATABASE_INIT_INFO)
	if err != nil {
		domain.Log(err)
		return nil, err
	}
	return handle, nil
}

func (handle *SqlHandler) Init() error {
	var info []struct {
		Name  string `db:"name"`
		Value string `db:"value"`
	}
	err := handle.db.Select(&info, "SELECT name, value FROM secret")
	if err != nil {
		return err
	}
	m := make(map[string]string)
	for _, v := range info {
		m[v.Name] = v.Value
	}
	domain.GMAIL_ADDRESS = m["gmailaddress"]
	domain.GMAIL_PASSWORD = m["gmailpassword"]
	return err
}

func (handle *SqlHandler) GetUserById(input_uuid *domain.UuidInt, input_tempid *domain.TempIdString) (*domain.SqlUser, error) {
	user := &domain.SqlUser{}
	err := handle.db.Get(user, "SELECT uuid, tempid, name, email, expire_tempid_at, send_email_count, send_last_email_at, email_authorized "+
		"FROM user WHERE uuid = ? AND tempid = ?",
		input_uuid.ToInt(), input_tempid.ToString(),
	)
	if err != nil {
		domain.Log(err)
		return nil, err
	}
	return user, nil
}

func (handle *SqlHandler) GetUser(input_name *domain.NameString, password *domain.PasswordString) (*domain.SqlUser, error) {
	user := &domain.SqlUser{}
	err := handle.db.Get(user, "SELECT uuid, tempid, name, email, expire_tempid_at, send_email_count, send_last_email_at, email_authorized, group_id "+
		"FROM user WHERE name=? AND password=?",
		input_name.ToString(), password.ToString(),
	)
	if err != nil {
		domain.Log(err)
		return nil, err
	}
	return user, nil
}

func (handle *SqlHandler) CreateUser(user *domain.User, password *domain.PasswordString) error {
	rows, err := handle.db.Queryx("SELECT COUNT(*) FROM user WHERE name=? UNION ALL "+
		"SELECT COUNT(*) FROM user WHERE email=? UNION ALL "+
		"SELECT COUNT(*) FROM user WHERE uuid=?",
		user.GetNameString(), password.ToString(), user.GetEmailString(),
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
		return err
	}
	if i != 3 {
		return domain.NewError("rows over 3")
	}
	res, err := handle.db.NamedExec("INSERT INTO user(uuid,tempid,name,password,email,expire_tempid_at,send_email_count,send_last_email_at, email_authorized, group_id) "+
		"VALUES (:uuid,:tempid,:name,:password,:email,:expire_tempid_at,:send_email_count,:send_last_email_at,:email_authorized,:group_id)", map[string]interface{}{
		"uuid":               user.GetUuidInt(),
		"tempid":             user.GetTempidString(),
		"name":               user.GetNameString(),
		"password":           password.ToString(),
		"email":              user.GetEmailString(),
		"expire_tempid_at":   user.GetTempidExpiredAt(),
		"send_email_count":   user.GetSendEmailCount(),
		"send_last_email_at": user.GetSendEmailLastDay(),
		"email_authorized":   user.EmailAuthorized(),
		"group_id":           user.GetGroupId(),
	})
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
		return domain.NewError("no inserted")
	}

	return nil
}

func (handle *SqlHandler) UpdateUser(user *domain.User) error {
	res, err := handle.db.NamedExec("UPDATE user SET tempid=:tempid,name=:name,email=:email,"+
		"expire_tempid_at=:expire_tempid_at,send_email_count=:send_email_count,send_last_email_at=:send_last_email_at,"+
		"email_authorized=:email_authorized, group_id=:group_id WHERE uuid=:uuid", map[string]interface{}{
		"uuid":               user.GetUuidInt(),
		"tempid":             user.GetTempidString(),
		"name":               user.GetNameString(),
		"email":              user.GetEmailString(),
		"expire_tempid_at":   user.GetTempidExpiredAt(),
		"send_email_count":   user.GetSendEmailCount(),
		"send_last_email_at": user.GetSendEmailLastDay(),
		"email_authorized":   user.EmailAuthorized(),
		"group_id":           user.GetGroupId(),
	})
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
		return domain.NewError("no updated")
	}
	return nil
}

func (handle *SqlHandler) DeleteForTest() {
	handle.db.Exec("DELETE FROM user WHERE email=?", domain.TEST_GMAIL_ADDRESS)
}
