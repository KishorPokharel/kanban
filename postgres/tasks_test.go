package postgres

import (
	"reflect"
	"testing"
)

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

func TestSortTaskInSameCategory(t *testing.T) {
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

	tasks := []*Task{
		{UserID: user.ID, Content: "A"},
		{UserID: user.ID, Content: "B"},
		{UserID: user.ID, Content: "C"},
	}
	want := []string{"B", "C", "A"}
	got := []string{}
	for _, task := range tasks {
		if err := service.Task.Insert(task); err != nil {
			t.Fatal(err)
		}
	}
	if err := service.Task.SortTaskInSameCategory(
		user.ID, tasks[0].ID, 0, 2, "TODO",
	); err != nil {
		t.Fatal(err)
	}
	allTasks, err := service.Task.GetAll(user.ID)
	if err != nil {
		t.Fatal(err)
	}
	val, ok := allTasks["TODO"]
	if ok {
		for _, t := range val {
			got = append(got, t.Content)
		}
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("sort in same category failed, got = %v, want = %v", got, want)
	}
}

func TestSortTaskInDifferentCategory(t *testing.T) {
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

	tasks := []*Task{
		{UserID: user.ID, Content: "A"},
		{UserID: user.ID, Content: "B"},
		{UserID: user.ID, Content: "C"},
		{UserID: user.ID, Content: "D"},
	}
	wantInTodo := []string{"A", "C", "D"}
	gotInTodo := []string{}
	wantInTesting := []string{"B"}
	gotInTesting := []string{}
	for _, task := range tasks {
		if err := service.Task.Insert(task); err != nil {
			t.Fatal(err)
		}
	}
	if err := service.Task.SortTaskInDifferentCategory(
		user.ID, tasks[1].ID, 1, 0, "TODO", "TESTING",
	); err != nil {
		t.Fatal(err)
	}
	allTasks, err := service.Task.GetAll(user.ID)
	if err != nil {
		t.Fatal(err)
	}
	val, ok := allTasks["TODO"]
	if ok {
		for _, t := range val {
			gotInTodo = append(gotInTodo, t.Content)
		}
	}
	if !reflect.DeepEqual(gotInTodo, wantInTodo) {
		t.Errorf("sort in different category failed, got = %v, want = %v", gotInTodo, wantInTodo)
	}
	val, ok = allTasks["TESTING"]
	if ok {
		for _, t := range val {
			gotInTesting = append(gotInTesting, t.Content)
		}
	}
	if !reflect.DeepEqual(gotInTesting, wantInTesting) {
		t.Errorf("sort in different category failed, got = %v, want = %v", gotInTesting, wantInTesting)
	}
}
