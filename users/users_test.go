package users

import (
	"os"
	"testing"

	"github.com/google/uuid"
)

func TestNewUserStore(t *testing.T) {
	tmpFile := "test_users.json"
	defer os.Remove(tmpFile)

	store, err := NewUserStore(tmpFile)
	if err != nil {
		t.Fatalf("Failed to create new user store: %v", err)
	}
	if store.Users == nil {
		t.Error("Users map not initialized")
	}
}

func TestUserStore_AddUser(t *testing.T) {
	tmpFile := "test_users.json"
	defer os.Remove(tmpFile)

	store, _ := NewUserStore(tmpFile)

	tests := []struct {
		name     string
		username string
		wantErr  bool
	}{
		{"valid user", "testuser", false},
		{"empty username", "", true},
		{"duplicate user", "testuser", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := store.AddUser(tt.username)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && user.UserName != tt.username {
				t.Errorf("AddUser() username = %v, want %v", user.UserName, tt.username)
			}
		})
	}
}

func TestUserStore_GetUser(t *testing.T) {
	tmpFile := "test_users.json"
	defer os.Remove(tmpFile)

	store, _ := NewUserStore(tmpFile)
	store.AddUser("testuser")

	tests := []struct {
		name     string
		username string
		want     bool
	}{
		{"existing user", "testuser", true},
		{"non-existing user", "nonexistent", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, got := store.GetUser(tt.username)
			if got != tt.want {
				t.Errorf("GetUser() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUserStore_GetUserById(t *testing.T) {
	tmpFile := "test_users.json"
	defer os.Remove(tmpFile)

	store, _ := NewUserStore(tmpFile)
	testUser, _ := store.AddUser("testuser")

	tests := []struct {
		name   string
		userId string
		want   bool
	}{
		{"existing user", testUser.UserId.String(), true},
		{"non-existing user", uuid.New().String(), false},
		{"invalid uuid", "invalid-uuid", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, got := store.GetUserById(tt.userId)
			if got != tt.want {
				t.Errorf("GetUserById() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUserStore_Save_Load(t *testing.T) {
	tmpFile := "test_users.json"
	defer os.Remove(tmpFile)

	store, _ := NewUserStore(tmpFile)
	testUser, _ := store.AddUser("testuser")

	// Test Save
	if err := store.Save(); err != nil {
		t.Errorf("Save() error = %v", err)
	}

	// Test Load
	newStore, _ := NewUserStore(tmpFile)
	if err := newStore.Load(); err != nil {
		t.Errorf("Load() error = %v", err)
	}

	if user, exists := newStore.GetUser("testuser"); !exists || user.UserName != testUser.UserName {
		t.Error("Load() failed to restore user data correctly")
	}
}
