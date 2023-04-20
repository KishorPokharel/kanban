package postgres

import "testing"

func TestInsertTask(t *testing.T) {
	db, tear := newTestDB(t)
	defer tear()
	service := NewService(db)

	pwd := "kishor123"
	user := &User{
		Username: "kishor",
		Email:    "kishor@gmail.com",
	}
	user.Password.Set(pwd)
	if err := service.User.Create(user); err != nil {
		t.Fatal(err)
	}

	task := &Task{
		UserID:  user.ID,
		Content: "Write Some Tests",
	}
	if err := service.Task.Insert(task); err != nil {
		t.Error(err)
	}
	if task.Category != "TODO" {
		t.Errorf("newly inserted task should have category TODO got = %s", task.Category)
	}
	if task.ID <= 0 {
		t.Errorf("task id should be > 0, got = %d", task.ID)
	}
}
