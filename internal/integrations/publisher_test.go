package integrations

import (
	"context"
	"errors"
	"testing"

	"github.com/juliofiliizzola/hookord/internal/domain"
)

type memoryRepo struct {
	mappings map[string]*domain.MessageMapping
	getErr   error
}

func (m *memoryRepo) GetMapping(_ context.Context, id string) (*domain.MessageMapping, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.mappings[id], nil
}

func (m *memoryRepo) SaveMapping(_ context.Context, mapping *domain.MessageMapping) error {
	if m.mappings == nil {
		m.mappings = map[string]*domain.MessageMapping{}
	}
	m.mappings[mapping.EntityID] = mapping
	return nil
}

func (m *memoryRepo) DeleteMapping(_ context.Context, id string) error {
	delete(m.mappings, id)
	return nil
}

type fakeMessenger struct {
	sent, edited []string
	lastEdit     MessageRef
	err          error
}

func (f *fakeMessenger) Send(channelID string, msg string) (MessageRef, error) {
	if f.err != nil {
		return MessageRef{}, f.err
	}
	f.sent = append(f.sent, msg)
	return MessageRef{ChannelID: channelID, MessageID: "m1"}, nil
}

func (f *fakeMessenger) Edit(ref MessageRef, msg string) (MessageRef, error) {
	if f.err != nil {
		return MessageRef{}, f.err
	}
	f.edited = append(f.edited, msg)
	f.lastEdit = ref
	return ref, nil
}

func TestPublish_SendsThenEdits(t *testing.T) {
	ctx := context.Background()
	repo := &memoryRepo{}
	messenger := &fakeMessenger{}
	entry := Entry{Integration: "fake", EntityID: 7, ChannelID: "c1", Repository: "org/repo", Status: "open"}

	if err := Publish(ctx, repo, messenger, entry, "v1"); err != nil {
		t.Fatalf("first publish: %v", err)
	}

	entry.Status = "closed"
	entry.TotalReviews = 2
	if err := Publish(ctx, repo, messenger, entry, "v2"); err != nil {
		t.Fatalf("second publish: %v", err)
	}

	if len(messenger.sent) != 1 || len(messenger.edited) != 1 {
		t.Fatalf("sent=%v edited=%v, want 1 send and 1 edit", messenger.sent, messenger.edited)
	}

	saved := repo.mappings["7-fake"]
	if saved == nil {
		t.Fatal("mapping not saved under key 7-fake")
	}
	if saved.ChannelID != "c1" || saved.MessageID != "m1" || saved.LastStatus != "closed" || saved.TotalReviews != 2 {
		t.Errorf("unexpected mapping: %+v", saved)
	}
	if saved.Repository != "org/repo" || saved.IntegrationName != "fake" {
		t.Errorf("unexpected mapping metadata: %+v", saved)
	}
}

func TestPublish_EditFallsBackToEntryChannel(t *testing.T) {
	repo := &memoryRepo{mappings: map[string]*domain.MessageMapping{
		"7-fake": {EntityID: "7-fake", MessageID: "m1"},
	}}
	messenger := &fakeMessenger{}

	entry := Entry{Integration: "fake", EntityID: 7, ChannelID: "c1"}
	if err := Publish(context.Background(), repo, messenger, entry, "v"); err != nil {
		t.Fatalf("publish: %v", err)
	}
	if messenger.lastEdit.ChannelID != "c1" {
		t.Errorf("edit channel = %q, want c1", messenger.lastEdit.ChannelID)
	}
}

func TestPublish_Errors(t *testing.T) {
	ctx := context.Background()
	entry := Entry{Integration: "fake", EntityID: 1}

	getErr := errors.New("redis down")
	if err := Publish(ctx, &memoryRepo{getErr: getErr}, &fakeMessenger{}, entry, "v"); !errors.Is(err, getErr) {
		t.Errorf("err = %v, want %v", err, getErr)
	}

	sendErr := errors.New("api down")
	repo := &memoryRepo{}
	if err := Publish(ctx, repo, &fakeMessenger{err: sendErr}, entry, "v"); !errors.Is(err, sendErr) {
		t.Errorf("err = %v, want %v", err, sendErr)
	}
	if len(repo.mappings) != 0 {
		t.Error("mapping must not be saved when send fails")
	}
}
