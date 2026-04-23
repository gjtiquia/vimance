package service_test

import (
	"testing"
)

func TestTagCRUD(t *testing.T) {
	s := setupTestService(t)

	user, _ := s.CreateUser(t.Context(), "testuser")

	tag, err := s.CreateTag(t.Context(), "food", "food expenses", "", user.ID)
	if err != nil {
		t.Fatalf("failed to create tag: %v", err)
	}

	if tag.ID == 0 {
		t.Error("expected tag ID to be set")
	}
	if tag.Name != "food" {
		t.Errorf("expected name 'food', got '%s'", tag.Name)
	}
	if tag.Description != "food expenses" {
		t.Errorf("expected description 'food expenses', got '%s'", tag.Description)
	}

	fetched, err := s.GetTag(t.Context(), tag.ID)
	if err != nil {
		t.Fatalf("failed to get tag: %v", err)
	}
	if fetched.Name != tag.Name {
		t.Errorf("expected name '%s', got '%s'", tag.Name, fetched.Name)
	}

	byName, err := s.GetTagByName(t.Context(), "food")
	if err != nil {
		t.Fatalf("failed to get tag by name: %v", err)
	}
	if byName.ID != tag.ID {
		t.Errorf("expected ID %d, got %d", tag.ID, byName.ID)
	}

	tags, err := s.ListTags(t.Context())
	if err != nil {
		t.Fatalf("failed to list tags: %v", err)
	}
	if len(tags) != 1 {
		t.Errorf("expected 1 tag, got %d", len(tags))
	}

	updated, err := s.UpdateTag(t.Context(), tag.ID, "groceries", "grocery shopping", "updated notes", user.ID)
	if err != nil {
		t.Fatalf("failed to update tag: %v", err)
	}
	if updated.Name != "groceries" {
		t.Errorf("expected name 'groceries', got '%s'", updated.Name)
	}
	if updated.Description != "grocery shopping" {
		t.Errorf("expected description 'grocery shopping', got '%s'", updated.Description)
	}

	err = s.SoftDeleteTag(t.Context(), tag.ID, user.ID)
	if err != nil {
		t.Fatalf("failed to soft delete tag: %v", err)
	}

	activeTags, err := s.ListActiveTags(t.Context())
	if err != nil {
		t.Fatalf("failed to list active tags: %v", err)
	}
	for _, at := range activeTags {
		if at.ID == tag.ID {
			t.Error("soft deleted tag should not appear in active_tags view")
		}
	}

	_, err = s.RestoreTag(t.Context(), tag.ID)
	if err != nil {
		t.Fatalf("failed to restore tag: %v", err)
	}

	activeTags, err = s.ListActiveTags(t.Context())
	if err != nil {
		t.Fatalf("failed to list active tags: %v", err)
	}
	found := false
	for _, at := range activeTags {
		if at.ID == tag.ID {
			found = true
			break
		}
	}
	if !found {
		t.Error("restored tag should appear in active_tags view")
	}

	err = s.HardDeleteTag(t.Context(), tag.ID)
	if err != nil {
		t.Fatalf("failed to hard delete tag: %v", err)
	}

	_, err = s.GetTag(t.Context(), tag.ID)
	if err == nil {
		t.Error("expected error when getting hard deleted tag")
	}
}
