package conference

import "context"

// Client — абстракция внешнего Conference Service
type Client interface {
	CreateLink(ctx context.Context) (url string, err error)
}
