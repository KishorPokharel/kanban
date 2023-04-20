package postgres

import "testing"

func TestUserCreate(t *testing.T) {
	tests := []struct {
		password  string
		user      *User
		wantError error
	}{
		{
			password: "kishor123",
			user: &User{
				Username: "kishor",
				Email:    "kishor@gmail.com",
			},
			wantError: nil,
		},
		{
			password: "bibek123",
			user: &User{
				Username: "bibek",
				Email:    "bibek@gmail.com",
			},
			wantError: nil,
		},
	}

	db, tear := newTestDB(t)
	defer tear()

	service := NewService(db)
	for _, tt := range tests {
		tt.user.Password.Set(tt.password)
		err := service.User.Create(tt.user)
		if err != tt.wantError {
			t.Errorf("want %v; got %s", tt.wantError, err)
		}
		if tt.user.ID <= 0 {
			t.Errorf("new user id <= 0, got id = %d", tt.user.ID)
		}
	}
}

func TestUserNoDuplicateEmail(t *testing.T) {
	tests := []struct {
		name      string
		password  string
		user      *User
		wantError error
	}{
		{
			name:     "New User should be created",
			password: "kishor123",
			user: &User{
				Username: "kishor",
				Email:    "kishor@gmail.com",
			},
			wantError: nil,
		},
		{
			name:     "Should error for duplicate email",
			password: "kishor123",
			user: &User{
				Username: "kishor",
				Email:    "kishor@gmail.com",
			},
			wantError: ErrDuplicateEmail,
		},
	}

	db, tear := newTestDB(t)
	defer tear()

	service := NewService(db)
	for _, tt := range tests {
		tt.user.Password.Set(tt.password)
		err := service.User.Create(tt.user)
		if err != tt.wantError {
			t.Errorf("want %v; got %s", tt.wantError, err)
		}
	}
}

func TestUserGetByEmail(t *testing.T) {
	tests := []struct {
		name      string
		password  string
		email     string
		user      *User
		wantError error
	}{
		{
			name:     "New User should be created",
			password: "kishor123",
			email:    "kishor@gmail.com",
			user: &User{
				Username: "kishor",
				Email:    "kishor@gmail.com",
			},
			wantError: nil,
		},
	}
	db, tear := newTestDB(t)
	defer tear()

	service := NewService(db)
	for _, tt := range tests {
		tt.user.Password.Set(tt.password)
		err := service.User.Create(tt.user)
		if err != tt.wantError {
			t.Errorf("want %v; got %s", tt.wantError, err)
		}
		if tt.user.ID <= 0 {
			t.Error("new user id <= 0")
		}
		user, err := service.User.GetByEmail(tt.email)
		if err != tt.wantError {
			t.Errorf("want %v; got %s", tt.wantError, err)
		}
		if user.Email != tt.user.Email {
			t.Errorf("want %v; got %s", tt.wantError, err)
		}
	}
}
