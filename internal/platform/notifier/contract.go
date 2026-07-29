package notifier

import "context"

type Notifier interface {
	Subscribe(channel string) Subscription
	Publish(ctx context.Context, channel, result string) error
}

type Subscription interface {
	Wait(ctx context.Context) (string, error)
	Close()
}
