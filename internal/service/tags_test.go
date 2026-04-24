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

func TestPinnedTags(t *testing.T) {
	s := setupTestService(t)

	user, _ := s.CreateUser(t.Context(), "testuser")

	tag1, _ := s.CreateTag(t.Context(), "food", "", "", user.ID)
	tag2, _ := s.CreateTag(t.Context(), "transport", "", "", user.ID)
	tag3, _ := s.CreateTag(t.Context(), "fun", "", "", user.ID)

	pinnedBefore, err := s.ListPinnedTags(t.Context())
	if err != nil {
		t.Fatalf("failed to list pinned tags: %v", err)
	}
	if len(pinnedBefore) != 0 {
		t.Errorf("expected 0 pinned tags, got %d", len(pinnedBefore))
	}

	err = s.PinTag(t.Context(), tag1.ID, user.ID)
	if err != nil {
		t.Fatalf("failed to pin tag: %v", err)
	}

	err = s.PinTag(t.Context(), tag3.ID, user.ID)
	if err != nil {
		t.Fatalf("failed to pin tag: %v", err)
	}

	pinned, err := s.ListPinnedTags(t.Context())
	if err != nil {
		t.Fatalf("failed to list pinned tags: %v", err)
	}
	if len(pinned) != 2 {
		t.Errorf("expected 2 pinned tags, got %d", len(pinned))
	}

	if pinned[0].Name != "food" {
		t.Errorf("expected first pinned tag 'food', got '%s'", pinned[0].Name)
	}
	if pinned[1].Name != "fun" {
		t.Errorf("expected second pinned tag 'fun', got '%s'", pinned[1].Name)
	}

	err = s.UnpinTag(t.Context(), tag1.ID)
	if err != nil {
		t.Fatalf("failed to unpin tag: %v", err)
	}

	pinnedAfter, err := s.ListPinnedTags(t.Context())
	if err != nil {
		t.Fatalf("failed to list pinned tags: %v", err)
	}
	if len(pinnedAfter) != 1 {
		t.Errorf("expected 1 pinned tag after unpin, got %d", len(pinnedAfter))
	}
	if pinnedAfter[0].Name != "fun" {
		t.Errorf("expected remaining pinned tag 'fun', got '%s'", pinnedAfter[0].Name)
	}

	err = s.PinTag(t.Context(), tag2.ID, user.ID)
	if err != nil {
		t.Fatalf("failed to pin tag2: %v", err)
	}

	pinnedFinal, err := s.ListPinnedTags(t.Context())
	if err != nil {
		t.Fatalf("failed to list pinned tags: %v", err)
	}
	if len(pinnedFinal) != 2 {
		t.Errorf("expected 2 pinned tags, got %d", len(pinnedFinal))
	}
	if pinnedFinal[1].Name != "transport" {
		t.Errorf("expected last pinned tag 'transport', got '%s'", pinnedFinal[1].Name)
	}
}
