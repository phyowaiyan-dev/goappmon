package repositories

import (
	"context"
	"testing"

	"github.com/phyowaiyan-dev/goappmon/internal/models"
)

func TestFeatureFlagRepositoryLifecycle(t *testing.T) {
	db := newTestDB(t)
	repo := NewFeatureFlagRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, "chat", true)
	if err != nil {
		t.Fatalf("create flag: %v", err)
	}
	if id == 0 {
		t.Fatal("expected non-zero id")
	}
	if _, err := repo.Create(ctx, "payment", false); err != nil {
		t.Fatalf("create second flag: %v", err)
	}

	flags, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("list flags: %v", err)
	}
	if len(flags) != 2 {
		t.Fatalf("expected 2 flags, got %d", len(flags))
	}

	flagMap, err := repo.AsMap(ctx)
	if err != nil {
		t.Fatalf("flags as map: %v", err)
	}
	if !flagMap["chat"] || flagMap["payment"] {
		t.Fatalf("unexpected flag map: %#v", flagMap)
	}

	if err := repo.Update(ctx, id, "chat-v2", false); err != nil {
		t.Fatalf("update flag: %v", err)
	}
	updated, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("list after update: %v", err)
	}
	if updated[0].Key != "chat-v2" || updated[0].Enabled {
		t.Fatalf("unexpected updated flag: %+v", updated[0])
	}

	if err := repo.Delete(ctx, id); err != nil {
		t.Fatalf("delete flag: %v", err)
	}
	remaining, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("list after delete: %v", err)
	}
	if len(remaining) != 1 {
		t.Fatalf("expected 1 flag, got %d", len(remaining))
	}
}

func TestFeatureFlagRepositoryMissing(t *testing.T) {
	db := newTestDB(t)
	repo := NewFeatureFlagRepository(db)
	ctx := context.Background()

	if err := repo.Update(ctx, 999, "missing", true); err != ErrFeatureFlagNotFound {
		t.Fatalf("expected ErrFeatureFlagNotFound on update, got %v", err)
	}
	if err := repo.Delete(ctx, 999); err != ErrFeatureFlagNotFound {
		t.Fatalf("expected ErrFeatureFlagNotFound on delete, got %v", err)
	}
	_ = models.FeatureFlag{}
}
