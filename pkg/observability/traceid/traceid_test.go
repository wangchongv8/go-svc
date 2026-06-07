package traceid

import (
	"context"
	"testing"

	"google.golang.org/grpc/metadata"
)

func TestFromContextReadsIncomingMetadata(t *testing.T) {
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
		traceIDMetadataKey, "trace-1",
		spanIDMetadataKey, "span-1",
	))

	tid, sid := FromContext(ctx)
	if tid != "trace-1" {
		t.Fatalf("trace id = %q, want trace-1", tid)
	}
	if sid != "span-1" {
		t.Fatalf("span id = %q, want span-1", sid)
	}
}

func TestWithOutgoingMetadataNoTraceLeavesContextUsable(t *testing.T) {
	ctx := WithOutgoingMetadata(context.Background())
	if ctx == nil {
		t.Fatal("context should not be nil")
	}
}
