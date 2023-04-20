package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"
)

var ErrInvalidData = errors.New("invalid task id or source index or destination index")

var categories = []string{"TODO", "IN PROGRESS", "TESTING", "DONE"}

type Task struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id,omitempty"`
	Category  string    `json:"category"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type TaskService struct {
	DB *sql.DB
}

func (ts TaskService) GetAll(userID int64) (map[string][]Task, error) {
	query := `
        select x.id, tasks.category, content, created_at 
        from taskorder, unnest(value)
        with ordinality as x(id)
        join tasks on tasks.id = x.id where tasks.user_id = $1;
    `
	rows, err := ts.DB.Query(query, userID)
	if err != nil {
		return map[string][]Task{}, err
	}

	tasks := map[string][]Task{}
	for _, val := range categories {
		tasks[val] = []Task{}
	}

	for rows.Next() {
		task := Task{}
		rows.Scan(&task.ID, &task.Category, &task.Content, &task.CreatedAt)
		_, ok := tasks[task.Category]
		if !ok {
			tasks[task.Category] = []Task{task}
			continue
		}
		tasks[task.Category] = append(tasks[task.Category], task)
	}
	if err := rows.Close(); err != nil {
		return map[string][]Task{}, err
	}
	return tasks, nil
}

func (ts TaskService) Insert(task *Task) error {
	queryInsertTask := `
        insert into tasks (user_id, content)
        values ($1, $2) returning id, category, created_at
    `
	args := []any{task.UserID, task.Content}
	tx, err := ts.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	taskRow := tx.QueryRowContext(context.Background(), queryInsertTask, args...)
	err = taskRow.Scan(&task.ID, &task.Category, &task.CreatedAt)
	if err != nil {
		return err
	}
	queryInsertOrder := `
        update taskorder set value = array_append(value, $1)
        where user_id = $2 and category = 'TODO';
    `
	_, err = tx.ExecContext(context.Background(), queryInsertOrder, task.ID, task.UserID)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (ts TaskService) SortTaskInSameCategory(userID, taskID, sourceIndex, destinationIndex int64, category string) error {
	query := `
        select value from taskorder
        where user_id = $1 and category = $2
    `
	row := ts.DB.QueryRow(query, userID, category)
	ids := []int64{}
	if err := row.Scan(pq.Array(&ids)); err != nil {
		return err
	}
	// TODO: Refactor
	idx, ok := taskIdInArray(taskID, ids)
	if !ok || idx != sourceIndex || destinationIndex > int64(len(ids)-1) {
		return ErrInvalidData
	}
	move(taskID, sourceIndex, destinationIndex, ids)
	queryUpdate := `
        update taskorder
        set value = $1
        where user_id = $2 and category = $3
    `
	args := []any{pq.Array(ids), userID, category}
	_, err := ts.DB.Exec(queryUpdate, args...)
	return err
}

func move(taskID, sourceIndex, destinationIndex int64, ids []int64) {
	if sourceIndex < destinationIndex {
		for i := sourceIndex; i < destinationIndex; i++ {
			ids[i] = ids[i+1]
		}
		ids[destinationIndex] = taskID
		return
	} else {
		for i := sourceIndex; i > destinationIndex; i-- {
			ids[i] = ids[i-1]
		}
		ids[destinationIndex] = taskID
		return
	}
}

func taskIdInArray(taskID int64, ids []int64) (int64, bool) {
	for idx, val := range ids {
		if val == taskID {
			return int64(idx), true
		}
	}
	return 0, false
}

func (ts TaskService) SortTaskInDifferentCategory(
	userID, taskID, sourceIndex, destinationIndex int64,
	sourceCategory, destinationCategory string,
) error {
	query := `
        select value from taskorder
        where user_id = $1 and category = $2
    `
	args := []any{userID, sourceCategory}
	tx, err := ts.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	row := tx.QueryRow(query, args...)
	sourceIDs := []int64{}
	if err := row.Scan(pq.Array(&sourceIDs)); err != nil {
		return err
	}
	idx, ok := taskIdInArray(taskID, sourceIDs)
	if !ok || idx != sourceIndex {
		return ErrInvalidData
	}

	newSourceIDs := []int64{}
	newSourceIDs = append(newSourceIDs, sourceIDs[:idx]...)
	newSourceIDs = append(newSourceIDs, sourceIDs[idx+1:]...)
	queryUpdate := `
        update taskorder
        set value = $1
        where user_id = $2 and category = $3
    `
	argsQueryUpdateSource := []any{pq.Array(&newSourceIDs), userID, sourceCategory}
	_, err = tx.Exec(queryUpdate, argsQueryUpdateSource...)
	if err != nil {
		return err
	}

	argsQueryDestination := []any{userID, destinationCategory}
	destinationRow := tx.QueryRow(query, argsQueryDestination...)
	destinationIDs := []int64{}
	if err := destinationRow.Scan(pq.Array(&destinationIDs)); err != nil {
		return err
	}
	if int(destinationIndex) > len(destinationIDs) {
		return ErrInvalidData
	}
	if int(destinationIndex) == len(destinationIDs) {
		destinationIDs = append(destinationIDs, taskID)
	} else {
		newDestinationIDs := []int64{}
		newDestinationIDs = append(newDestinationIDs, destinationIDs...)
		destinationIDs = destinationIDs[:destinationIndex]
		destinationIDs = append(destinationIDs, taskID)
		destinationIDs = append(destinationIDs, newDestinationIDs[destinationIndex:]...)
	}

	argsQueryUpdateDestination := []any{pq.Array(&destinationIDs), userID, destinationCategory}
	_, err = tx.Exec(queryUpdate, argsQueryUpdateDestination...)
	if err != nil {
		return err
	}

	queryUpdateTasks := `
        update tasks
        set category = $1
        where id = $2
    `
	_, err = tx.Exec(queryUpdateTasks, destinationCategory, taskID)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}
