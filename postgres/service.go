package postgres

import "database/sql"

type Service struct {
	User  UserService
	Token TokenService
	Task  TaskService
}

func NewService(db *sql.DB) Service {
	s := Service{
		User:  UserService{DB: db},
		Token: TokenService{DB: db},
		Task:  TaskService{DB: db},
	}
	return s
}
