package internal

import (
	"fmt"
	"testing"
	"time"
)

func TestNewTaskHolder(t *testing.T) {
	t.Run("Creates a new TaskHolder", func(t *testing.T) {
		th := NewTaskHolder("resources/cli_disk_test.json")
		if th == nil {
			t.Errorf("Expected NewTaskHolder to return a non-nil pointer")
		}

		if th.latestId != 0 {
			t.Errorf("got %v, want 0", th.latestId)
		}

		if len(th.Tasks) != 0 {
			t.Errorf("Expected tasks to be empty, got %d tasks", len(th.Tasks))
		}
	})
}

func TestAddTask(t *testing.T) {
	t.Run("Add a task", func(t *testing.T) {
		th := NewTaskHolder("resources/cli_disk_test.json")
		task := ProvideTask(t)
		th.Add(task)
		if len(th.Tasks) != 1 {
			t.Errorf("Expected tasks to be 1, got %d tasks", len(th.Tasks))
		}
		if th.latestId != 1 {
			t.Errorf("got %v, want 1", th.latestId)

		}
	})
}

func TestReadTask(t *testing.T) {
	t.Run("Read Tasks", func(t *testing.T) {
		th := NewTaskHolder("resources/cli_disk_test.json")
		testTask := ProvideTask(t)
		th.Add(testTask)
		allTasks := th.Read()
		if len(allTasks) != 1 {
			t.Errorf("got %d want 1", len(allTasks))
		}
		if allTasks[0] != testTask {
			t.Errorf("got %v want %v", allTasks[0], testTask)
		}
	})
}

func TestCreateTask(t *testing.T) {
	th := NewTaskHolder("resources/cli_disk_test.json")
	taskValue := "Test task"
	category := TaskCategory(1)
	fmt.Println(category)
	plannedAt := time.Now()

	task := th.CreateTask(taskValue, category, plannedAt)

	if task.Id != 1 {
		t.Errorf("Expected task ID to be 1, got %d", task.Id)
	}

	if task.Msg != taskValue {
		t.Errorf("Expected task value to be %q, got %q", taskValue, task.Msg)
	}

	if task.Category != category {
		t.Errorf("Expected task category to be %q, got %q", category, task.Category)
	}

	if !task.PlannedAt.Equal(plannedAt) {
		t.Errorf("Expected plannedAt to be %v, got %v", plannedAt, task.PlannedAt)
	}

	if len(th.Tasks) != 1 {
		t.Errorf("Expected TaskHolder to have 1 task, got %d", len(th.Tasks))
	}

	if th.latestId != 1 {
		t.Errorf("Expected latestId to be 1, got %d", th.latestId)
	}
}

func TestFindTaskById(t *testing.T) {
	th := NewTaskHolder("resources/cli_disk_test.json")
	task1 := th.CreateTask("Task 1", TaskCategory(1), time.Now().Add(time.Minute))
	th.CreateTask("Task 2", TaskCategory(1), time.Now())

	t.Run("Find existing task", func(t *testing.T) {
		foundTask, err := th.FindTaskById(task1.Id)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if foundTask == nil {
			t.Fatal("Expected to find a task, got nil")
		}
		if foundTask.Id != task1.Id {
			t.Errorf("Expected task ID %d, got %d", task1.Id, foundTask.Id)
		}
	})

	t.Run("Find non-existent task", func(t *testing.T) {
		nonExistentId := 999
		_, err := th.FindTaskById(nonExistentId)
		if err != ErrNotFound {
			t.Errorf("Expected ErrNotFound, got %v", err)
		}
	})
}

func TestPartialUpdateTask(t *testing.T) {
	setupTest := func() (*TaskHolder, *Task) {
		th := NewTaskHolder("resources/cli_disk_test.json")
		initialTask := th.CreateTask("Initial task", TaskCategory(0), time.Now().Add(24*time.Hour))
		return th, initialTask
	}

	t.Run("Update Done status", func(t *testing.T) {
		th, initialTask := setupTest()

		done := true

		// do update
		err := th.PartialUpdateTask(initialTask.Id, TaskUpdate{Done: &done})

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		updatedTask, _ := th.FindTaskById(initialTask.Id)

		// Check changed
		if !updatedTask.Done {
			t.Errorf("Expected task to be done, but it's not")
		}

		// Check that other fields haven't changed
		assertUnchanged(updatedTask, initialTask, t, []string{"msg", "category", "planned", "created", "id"})
	})

	t.Run("Update Msg", func(t *testing.T) {
		th, initialTask := setupTest()

		newMsg := "Updated task message"
		err := th.PartialUpdateTask(initialTask.Id, TaskUpdate{Msg: &newMsg})
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		updatedTask, _ := th.FindTaskById(initialTask.Id)
		if updatedTask.Msg != newMsg {
			t.Errorf("Expected task message to be %q, got %q", newMsg, updatedTask.Msg)
		}
		assertUnchanged(updatedTask, initialTask, t, []string{"done", "category", "planned", "created", "id"})

	})

	t.Run("Update with empty Msg", func(t *testing.T) {
		th, initialTask := setupTest()

		emptyMsg := ""
		err := th.PartialUpdateTask(initialTask.Id, TaskUpdate{Msg: &emptyMsg})
		if _, ok := err.(*EmptyTaskValueError); !ok {
			t.Errorf("Expected EmptyTaskValueError, got %v", err)
		}

		updatedTask, _ := th.FindTaskById(initialTask.Id)
		assertUnchanged(updatedTask, initialTask, t, []string{"done", "msg", "category", "planned", "created", "id"})

	})

	t.Run("Update Category", func(t *testing.T) {
		th, initialTask := setupTest()

		newCategory := TaskCategory(2)
		err := th.PartialUpdateTask(initialTask.Id, TaskUpdate{Category: &newCategory})
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		updatedTask, _ := th.FindTaskById(initialTask.Id)
		if updatedTask.Category != newCategory {
			t.Errorf("Expected task category to be %q, got %q", newCategory, updatedTask.Category)
		}
		assertUnchanged(updatedTask, initialTask, t, []string{"done", "msg", "planned", "created", "id"})

	})

	t.Run("Update with invalid Category", func(t *testing.T) {
		th, initialTask := setupTest()

		invalidCategory := TaskCategory(999)
		err := th.PartialUpdateTask(initialTask.Id, TaskUpdate{Category: &invalidCategory})
		if _, ok := err.(*InvalidCategoryError); !ok {
			t.Errorf("Expected InvalidCategoryError, got %v", err)
		}
		updatedTask, _ := th.FindTaskById(initialTask.Id)
		assertUnchanged(updatedTask, initialTask, t, []string{"done", "msg", "category", "planned", "created", "id"})

	})

	t.Run("Update PlannedAt", func(t *testing.T) {
		th, initialTask := setupTest()

		newPlannedAt := time.Now().Add(48 * time.Hour)
		err := th.PartialUpdateTask(initialTask.Id, TaskUpdate{PlannedAt: &newPlannedAt})
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		updatedTask, _ := th.FindTaskById(initialTask.Id)
		if !updatedTask.PlannedAt.Equal(newPlannedAt) {
			t.Errorf("Expected planned time to be %v, got %v", newPlannedAt, updatedTask.PlannedAt)
		}
		assertUnchanged(updatedTask, initialTask, t, []string{"done", "msg", "category", "done", "created", "id"})

	})

	t.Run("Update with past PlannedAt", func(t *testing.T) {
		th, initialTask := setupTest()

		pastTime := time.Now().Add(-24 * time.Hour)
		err := th.PartialUpdateTask(initialTask.Id, TaskUpdate{PlannedAt: &pastTime})
		if _, ok := err.(*PastPlannedTimeError); !ok {
			t.Errorf("Expected PastPlannedTimeError, got %v", err)
		}
	})

	t.Run("Update non-existent task", func(t *testing.T) {
		th, _ := setupTest()

		nonExistentId := 999
		err := th.PartialUpdateTask(nonExistentId, TaskUpdate{})
		if err != ErrNotFound {
			t.Errorf("Expected ErrNotFound, got %v", err)
		}
	})

	t.Run("Full update", func(t *testing.T) {
		th, initialTask := setupTest()

		newDone := true
		newMsg := "Fully updated task"
		newCategory := TaskCategory(3)
		newPlannedAt := time.Now().Add(72 * time.Hour)

		update := TaskUpdate{
			Done:      &newDone,
			Msg:       &newMsg,
			Category:  &newCategory,
			PlannedAt: &newPlannedAt,
		}

		err := th.PartialUpdateTask(initialTask.Id, update)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		updatedTask, _ := th.FindTaskById(initialTask.Id)

		if updatedTask.Done != newDone {
			t.Errorf("Expected Done to be %v, got %v", newDone, updatedTask.Done)
		}
		if updatedTask.Msg != newMsg {
			t.Errorf("Expected Msg to be %q, got %q", newMsg, updatedTask.Msg)
		}
		if updatedTask.Category != newCategory {
			t.Errorf("Expected Category to be %q, got %q", newCategory, updatedTask.Category)
		}
		if !updatedTask.PlannedAt.Equal(newPlannedAt) {
			t.Errorf("Expected PlannedAt to be %v, got %v", newPlannedAt, updatedTask.PlannedAt)
		}
		assertUnchanged(updatedTask, initialTask, t, []string{"created", "id"})

	})
}

func TestDeleteTask(t *testing.T) {
	t.Run("Delete task", func(t *testing.T) {
		th := NewTaskHolder("resources/cli_disk_test.json")
		task1 := th.CreateTask("Task 1", TaskCategory(1), time.Now().Add(time.Minute))
		th.CreateTask("Task 2", TaskCategory(1), time.Now())

		err := th.DeleteTask(task1.Id)
		if err != nil {
			t.Errorf("Unexpected error ocuured")
		}

		if len(th.Tasks) != 1 {
			t.Errorf("got 1 want %d", len(th.Tasks))
		}

		if th.Read()[0].Msg != "Task 2" {
			t.Errorf("Expected task msg %q, but gor %q", "Task2", th.Read()[0].Msg)
		}
	})

	t.Run("Delete task with wrong id", func(t *testing.T) {
		th := NewTaskHolder("resources/cli_disk_test.json")
		th.CreateTask("Task 1", TaskCategory(1), time.Now().Add(time.Minute))
		th.CreateTask("Task 2", TaskCategory(1), time.Now())

		err := th.DeleteTask(9999)
		if err == nil {
			t.Errorf("Expected error, got nil")
		}
	})
}

func assertUnchanged(updatedTask *Task, initialTask *Task, t *testing.T, fields []string) {
	for _, field := range fields {
		if field == "msg" {
			if updatedTask.Msg != initialTask.Msg {
				t.Errorf("Expected Msg to remain %q, but got %q", initialTask.Msg, updatedTask.Msg)
			}
		}
		if field == "category" {
			if updatedTask.Category != initialTask.Category {
				t.Errorf("Expected Category to remain %q, but got %q", initialTask.Category, updatedTask.Category)
			}
		}
		if field == "planned" {
			if !updatedTask.PlannedAt.Equal(initialTask.PlannedAt) {
				t.Errorf("Expected PlannedAt to remain %v, but got %v", initialTask.PlannedAt, updatedTask.PlannedAt)
			}
		}
		if field == "created" {
			if !updatedTask.CreatedAt.Equal(initialTask.CreatedAt) {
				t.Errorf("Expected CreatedAt to remain %v, but got %v", initialTask.CreatedAt, updatedTask.CreatedAt)
			}
		}
		if field == "id" {

			if !(updatedTask.Id == initialTask.Id) {
				t.Errorf("Expected Id to remain %v, but got %v", initialTask.Id, updatedTask.Id)
			}
		}
		if field == "done" {
			if !(updatedTask.Done == initialTask.Done) {
				t.Errorf("Expected Done to remain %v, but got %v", initialTask.Done, updatedTask.Done)
			}
		}
	}
}
