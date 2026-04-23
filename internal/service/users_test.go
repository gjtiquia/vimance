package service_test

import (
	"testing"
)

func TestUserCRUD(t *testing.T) {
	s := setupTestService(t)

	user, err := s.CreateUser(t.Context(), "testuser")
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	if user.ID == 0 {
		t.Error("expected user ID to be set")
	}
	if user.Username != "testuser" {
		t.Errorf("expected username 'testuser', got '%s'", user.Username)
	}

	fetched, err := s.GetUser(t.Context(), user.ID)
	if err != nil {
		t.Fatalf("failed to get user: %v", err)
	}
	if fetched.Username != user.Username {
		t.Errorf("expected username '%s', got '%s'", user.Username, fetched.Username)
	}

	err = s.SoftDeleteUser(t.Context(), user.ID, user.ID)
	if err != nil {
		t.Fatalf("failed to soft delete user: %v", err)
	}

	activeUsers, err := s.ListActiveUsers(t.Context())
	if err != nil {
		t.Fatalf("failed to list active users: %v", err)
	}
	for _, u := range activeUsers {
		if u.ID == user.ID {
			t.Error("soft deleted user should not appear in active_users view")
		}
	}

	_, err = s.RestoreUser(t.Context(), user.ID)
	if err != nil {
		t.Fatalf("failed to restore user: %v", err)
	}

	activeUsers, err = s.ListActiveUsers(t.Context())
	if err != nil {
		t.Fatalf("failed to list active users: %v", err)
	}
	found := false
	for _, u := range activeUsers {
		if u.ID == user.ID {
			found = true
			break
		}
	}
	if !found {
		t.Error("restored user should appear in active_users view")
	}

	updated, err := s.UpdateUser(t.Context(), user.ID, "updateduser")
	if err != nil {
		t.Fatalf("failed to update user: %v", err)
	}
	if updated.Username != "updateduser" {
		t.Errorf("expected username 'updateduser', got '%s'", updated.Username)
	}

	err = s.HardDeleteUser(t.Context(), user.ID)
	if err != nil {
		t.Fatalf("failed to hard delete user: %v", err)
	}

	_, err = s.GetUser(t.Context(), user.ID)
	if err == nil {
		t.Error("expected error when getting hard deleted user")
	}
}
