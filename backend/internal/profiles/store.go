package profiles

import "context"

type Store interface {
    List(ctx context.Context) ([]Profile, error)
    Get(ctx context.Context, id string) (Profile, error)
    Save(ctx context.Context, profile Profile) (string, error)
    Delete(ctx context.Context, id string) error
}
