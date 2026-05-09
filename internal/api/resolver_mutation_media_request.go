package api

import "context"

// Galileo fork: mutation resolvers for the media request workflow.
// Stubs only; persistence + Service wiring lands in a follow-up commit.

func (r *mutationResolver) MediaRequestCreate(ctx context.Context, input MediaRequestCreateInput) (*MediaRequest, error) {
	return nil, errMediaRequestsNotWired
}

func (r *mutationResolver) MediaRequestSearch(ctx context.Context, id string) (*MediaRequest, error) {
	return nil, errMediaRequestsNotWired
}

func (r *mutationResolver) MediaRequestApprove(ctx context.Context, releaseID string) (*MediaRequest, error) {
	return nil, errMediaRequestsNotWired
}

func (r *mutationResolver) MediaRequestReject(ctx context.Context, id string, reason *string) (*MediaRequest, error) {
	return nil, errMediaRequestsNotWired
}

func (r *mutationResolver) MediaRequestDestroy(ctx context.Context, id string) (bool, error) {
	return false, errMediaRequestsNotWired
}
