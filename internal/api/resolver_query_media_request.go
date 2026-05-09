package api

import (
	"context"
	"errors"
)

// Galileo fork: query resolvers for the media request workflow.
// Wiring to the persistence layer + Service lands in a follow-up commit;
// for now these stubs let the generated GraphQL schema compile.

var errMediaRequestsNotWired = errors.New("media requests: persistence + service not wired yet")

func (r *queryResolver) FindMediaRequest(ctx context.Context, id string) (*MediaRequest, error) {
	return nil, errMediaRequestsNotWired
}

func (r *queryResolver) FindMediaRequests(ctx context.Context, filter *MediaRequestFilter) ([]*MediaRequest, error) {
	return nil, errMediaRequestsNotWired
}
