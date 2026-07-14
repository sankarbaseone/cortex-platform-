package nyxbus

import (
	"context"
	"errors"
	"testing"

	commonv1 "github.com/nydux/platform/api/nydux/common/v1"
	"github.com/nydux/platform/libs/go/nyxpg"
)

type fakeReader struct {
	rows       []nyxpg.OutboxRecord
	published  []int64
	pendingErr error
}

func (f *fakeReader) Pending(_ context.Context, limit int) ([]nyxpg.OutboxRecord, error) {
	if f.pendingErr != nil {
		return nil, f.pendingErr
	}
	if limit < len(f.rows) {
		return f.rows[:limit], nil
	}
	return f.rows, nil
}

func (f *fakeReader) MarkPublished(_ context.Context, ids []int64) error {
	f.published = append(f.published, ids...)
	return nil
}

type fakePublisher struct {
	published []*commonv1.EventEnvelope
	failAfter int // fail starting at this call index (0 = never fails)
	calls     int
}

func (f *fakePublisher) Publish(_ context.Context, _ string, env *commonv1.EventEnvelope) error {
	f.calls++
	if f.failAfter > 0 && f.calls > f.failAfter {
		return errors.New("kafka unavailable")
	}
	f.published = append(f.published, env)
	return nil
}

func TestRelayOnce_PublishesAllPendingAndMarksThem(t *testing.T) {
	reader := &fakeReader{rows: []nyxpg.OutboxRecord{
		{ID: 1, Tenant: "t1", EventType: "nydux.compiler.kernel.scored.v1", PartitionKey: "t1:hash1", Payload: []byte("a")},
		{ID: 2, Tenant: "t1", EventType: "nydux.compiler.kernel.scored.v1", PartitionKey: "t1:hash2", Payload: []byte("b")},
	}}
	pub := &fakePublisher{}
	o := &OutboxPublisher{reader: reader, producer: pub, batchSize: 500}

	if err := o.relayOnce(context.Background()); err != nil {
		t.Fatalf("relayOnce() error = %v", err)
	}
	if len(pub.published) != 2 {
		t.Fatalf("published %d envelopes, want 2", len(pub.published))
	}
	if len(reader.published) != 2 || reader.published[0] != 1 || reader.published[1] != 2 {
		t.Fatalf("MarkPublished ids = %v, want [1 2]", reader.published)
	}
}

func TestRelayOnce_NoPendingRowsIsNoop(t *testing.T) {
	reader := &fakeReader{}
	pub := &fakePublisher{}
	o := &OutboxPublisher{reader: reader, producer: pub, batchSize: 500}

	if err := o.relayOnce(context.Background()); err != nil {
		t.Fatalf("relayOnce() error = %v", err)
	}
	if len(pub.published) != 0 || len(reader.published) != 0 {
		t.Fatalf("expected no publish/mark calls on empty outbox")
	}
}

func TestRelayOnce_StopsAtFirstPublishFailure_ResumableNextTick(t *testing.T) {
	reader := &fakeReader{rows: []nyxpg.OutboxRecord{
		{ID: 1, EventType: "e", PartitionKey: "k1", Payload: []byte("a")},
		{ID: 2, EventType: "e", PartitionKey: "k2", Payload: []byte("b")},
		{ID: 3, EventType: "e", PartitionKey: "k3", Payload: []byte("c")},
	}}
	pub := &fakePublisher{failAfter: 1} // first Publish succeeds, second fails
	o := &OutboxPublisher{reader: reader, producer: pub, batchSize: 500}

	if err := o.relayOnce(context.Background()); err != nil {
		t.Fatalf("relayOnce() error = %v", err)
	}
	if len(pub.published) != 1 {
		t.Fatalf("published %d envelopes before failure, want 1", len(pub.published))
	}
	if len(reader.published) != 1 || reader.published[0] != 1 {
		t.Fatalf("MarkPublished must only cover the successfully published row, got %v", reader.published)
	}
}

func TestRelayOnce_PendingErrorPropagates(t *testing.T) {
	reader := &fakeReader{pendingErr: errors.New("pg down")}
	o := &OutboxPublisher{reader: reader, producer: &fakePublisher{}, batchSize: 500}
	if err := o.relayOnce(context.Background()); err == nil {
		t.Fatal("expected Pending error to propagate")
	}
}
