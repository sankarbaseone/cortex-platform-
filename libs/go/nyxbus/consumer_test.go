package nyxbus

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
	"google.golang.org/protobuf/proto"

	commonv1 "github.com/nydux/platform/api/nydux/common/v1"
)

type fakeRawPublisher struct {
	calls []struct {
		topic      string
		key, value []byte
	}
}

func (f *fakeRawPublisher) publishRaw(_ context.Context, topic string, key, value []byte) error {
	f.calls = append(f.calls, struct {
		topic      string
		key, value []byte
	}{topic, key, value})
	return nil
}

func envelopeBytes(t *testing.T, env *commonv1.EventEnvelope) []byte {
	t.Helper()
	b, err := proto.Marshal(env)
	if err != nil {
		t.Fatalf("marshal envelope: %v", err)
	}
	return b
}

func TestProcessRecord_UndecodableEnvelopeGoesStraightToDLQ(t *testing.T) {
	dlq := &fakeRawPublisher{}
	c := &Consumer{
		cfg: ConsumerConfig{Topic: "compiler.kernel.scored"},
		dlq: dlq,
		handler: func(context.Context, *commonv1.EventEnvelope) error {
			t.Fatal("handler must not run for undecodable envelope")
			return nil
		},
	}
	c.cfg.setDefaults()
	rec := &kgo.Record{Key: []byte("k"), Value: []byte("not a valid protobuf envelope \xff\xfe")}
	c.processRecord(context.Background(), rec)

	if len(dlq.calls) != 1 || dlq.calls[0].topic != "compiler.kernel.scored.dlq" {
		t.Fatalf("expected one DLQ publish to compiler.kernel.scored.dlq, got %+v", dlq.calls)
	}
}

func TestProcessRecord_PoisonGoesStraightToDLQNoRetry(t *testing.T) {
	dlq := &fakeRawPublisher{}
	calls := 0
	c := &Consumer{
		cfg: ConsumerConfig{Topic: "t"},
		dlq: dlq,
		handler: func(context.Context, *commonv1.EventEnvelope) error {
			calls++
			return Poison(errors.New("bad schema version"))
		},
	}
	c.cfg.setDefaults()
	rec := &kgo.Record{Key: []byte("k"), Value: envelopeBytes(t, &commonv1.EventEnvelope{EventId: "e1"})}
	c.processRecord(context.Background(), rec)

	if calls != 1 {
		t.Fatalf("handler called %d times, want exactly 1 (no retry on poison)", calls)
	}
	if len(dlq.calls) != 1 || dlq.calls[0].topic != "t.dlq" {
		t.Fatalf("expected one DLQ publish to t.dlq, got %+v", dlq.calls)
	}
}

func TestProcessRecord_SucceedsOnRetryAfterTransientError(t *testing.T) {
	dlq := &fakeRawPublisher{}
	calls := 0
	c := &Consumer{
		cfg: ConsumerConfig{Topic: "t", MaxRetries: 3, BaseBackoff: time.Millisecond, MaxBackoff: 5 * time.Millisecond},
		dlq: dlq,
		handler: func(context.Context, *commonv1.EventEnvelope) error {
			calls++
			if calls < 3 {
				return errors.New("transient pg error")
			}
			return nil
		},
	}
	rec := &kgo.Record{Key: []byte("k"), Value: envelopeBytes(t, &commonv1.EventEnvelope{EventId: "e1"})}
	c.processRecord(context.Background(), rec)

	if calls != 3 {
		t.Fatalf("handler called %d times, want 3 (2 failures + 1 success)", calls)
	}
	if len(dlq.calls) != 0 {
		t.Fatalf("expected no DLQ/.retry publish on eventual success, got %+v", dlq.calls)
	}
}

func TestProcessRecord_ExhaustedRetriesGoToRetryTopic(t *testing.T) {
	dlq := &fakeRawPublisher{}
	c := &Consumer{
		cfg: ConsumerConfig{Topic: "t", MaxRetries: 2, BaseBackoff: time.Millisecond, MaxBackoff: 2 * time.Millisecond},
		dlq: dlq,
		handler: func(context.Context, *commonv1.EventEnvelope) error {
			return errors.New("still down")
		},
	}
	rec := &kgo.Record{Key: []byte("k"), Value: envelopeBytes(t, &commonv1.EventEnvelope{EventId: "e1"})}
	c.processRecord(context.Background(), rec)

	if len(dlq.calls) != 1 || dlq.calls[0].topic != "t.retry" {
		t.Fatalf("expected exactly one publish to t.retry after exhausting retries, got %+v", dlq.calls)
	}
}

func TestJittered_NeverExceedsInput(t *testing.T) {
	d := 100 * time.Millisecond
	for i := 0; i < 100; i++ {
		if got := jittered(d); got < 0 || got > d {
			t.Fatalf("jittered(%v) = %v, out of [0,%v] range", d, got, d)
		}
	}
}

func TestJittered_ZeroIsZero(t *testing.T) {
	if got := jittered(0); got != 0 {
		t.Fatalf("jittered(0) = %v, want 0", got)
	}
}
