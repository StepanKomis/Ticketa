package migrations

import (
	"database/sql"
	_ "embed"

	migrate "github.com/StepanKomis/Ticketa/src/database/migrations"
)

//go:embed up/UP_00001.sql
var up00001 string

//go:embed up/UP_00002.sql
var up00002 string

var All = func() []migrate.Migration {
	ms := []migrate.Migration{
		{
			Name: "create_users_and_auth_tables",
			Up: func(db any) error {
				_, err := db.(*sql.DB).Exec(up00001)
				return err
			},
			Down: func(db any) error {
				_, err := db.(*sql.DB).Exec(`
					DROP TABLE IF EXISTS maintainer_profile, staff_profile, student_profile, ldap_login, local_login, users CASCADE;
					DROP TYPE IF EXISTS user_type, auth_provider;
				`)
				return err
			},
		},
		{
			Name: "create_sessions",
			Up: func(db any) error {
				_, err := db.(*sql.DB).Exec(up00002)
				return err
			},
			Down: func(db any) error {
				_, err := db.(*sql.DB).Exec(`DROP TABLE IF EXISTS sessions CASCADE;`)
				return err
			},
		},
	}
	for i := range ms {
		ms[i].Number = i + 1
	}
	return ms
}()
