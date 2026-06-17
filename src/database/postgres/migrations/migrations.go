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

//go:embed up/UP_00003.sql
var up00003 string

//go:embed up/UP_00004.sql
var up00004 string

//go:embed up/UP_00005.sql
var up00005 string

//go:embed up/UP_00006.sql
var up00006 string

//go:embed up/UP_00007.sql
var up00007 string

//go:embed up/UP_00008.sql
var up00008 string

//go:embed up/UP_00009.sql
var up00009 string

//go:embed up/UP_00010.sql
var up00010 string

//go:embed up/UP_00011.sql
var up00011 string

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
		{
			Name: "create_ticket_statuses",
			Up: func(db any) error {
				_, err := db.(*sql.DB).Exec(up00003)
				return err
			},
			Down: func(db any) error {
				_, err := db.(*sql.DB).Exec(`DROP TABLE IF EXISTS ticket_statuses CASCADE;`)
				return err
			},
		},
		{
			Name: "create_tickets",
			Up: func(db any) error {
				_, err := db.(*sql.DB).Exec(up00004)
				return err
			},
			Down: func(db any) error {
				_, err := db.(*sql.DB).Exec(`DROP TABLE IF EXISTS tickets CASCADE;`)
				return err
			},
		},
		{
			Name: "create_ticket_comments",
			Up: func(db any) error {
				_, err := db.(*sql.DB).Exec(up00005)
				return err
			},
			Down: func(db any) error {
				_, err := db.(*sql.DB).Exec(`DROP TABLE IF EXISTS ticket_comments CASCADE;`)
				return err
			},
		},
		{
			Name: "add_pending_and_admin_user_types",
			Up: func(db any) error {
				_, err := db.(*sql.DB).Exec(up00006)
				return err
			},
			// PostgreSQL enum values cannot be removed; down is a no-op.
			Down: func(db any) error { return nil },
		},
		{
			Name: "add_requested_role_and_approved_by",
			Up: func(db any) error {
				_, err := db.(*sql.DB).Exec(up00007)
				return err
			},
			Down: func(db any) error {
				_, err := db.(*sql.DB).Exec(`
					ALTER TABLE users DROP COLUMN IF EXISTS requested_role;
					ALTER TABLE users DROP COLUMN IF EXISTS approved_by;
				`)
				return err
			},
		},
		{
			Name: "create_invitations",
			Up: func(db any) error {
				_, err := db.(*sql.DB).Exec(up00008)
				return err
			},
			Down: func(db any) error {
				_, err := db.(*sql.DB).Exec(`DROP TABLE IF EXISTS invitations CASCADE;`)
				return err
			},
		},
		{
			Name: "ticket_improvements_priority_votes",
			Up: func(db any) error {
				_, err := db.(*sql.DB).Exec(up00009)
				return err
			},
			Down: func(db any) error {
				_, err := db.(*sql.DB).Exec(`
					DROP TABLE IF EXISTS ticket_votes CASCADE;
					DROP TRIGGER IF EXISTS trg_ticket_updated_at ON tickets;
					DROP FUNCTION IF EXISTS set_ticket_updated_at;
					ALTER TABLE tickets
						DROP COLUMN IF EXISTS priority,
						DROP COLUMN IF EXISTS assigned_to,
						DROP COLUMN IF EXISTS location,
						DROP COLUMN IF EXISTS category,
						DROP COLUMN IF EXISTS updated_at;
				`)
				return err
			},
		},
		{
			Name: "ticket_history",
			Up: func(db any) error {
				_, err := db.(*sql.DB).Exec(up00010)
				return err
			},
			Down: func(db any) error {
				_, err := db.(*sql.DB).Exec(`DROP TABLE IF EXISTS ticket_history CASCADE;`)
				return err
			},
		},
		{
			Name: "activity_log",
			Up: func(db any) error {
				_, err := db.(*sql.DB).Exec(up00011)
				return err
			},
			Down: func(db any) error {
				_, err := db.(*sql.DB).Exec(`DROP TABLE IF EXISTS activity_log CASCADE;`)
				return err
			},
		},
	}
	for i := range ms {
		ms[i].Number = i + 1
	}
	return ms
}()
