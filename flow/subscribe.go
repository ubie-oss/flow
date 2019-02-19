package flow

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"

	"cloud.google.com/go/pubsub"
	"github.com/sakajunquality/cloud-pubsub-events/cloudbuildevent"
)

var (
	mu     sync.Mutex
	killCh chan bool
)

func (f *Flow) subscribeGCB(ctx context.Context, errCh chan error) {
	killCh := make(chan bool, 2)

	for {
		select {
		case <-ctx.Done():
			errCh <- ctx.Err()
		default:
		}

		if len(killCh) > 0 {
			break
		}

		err := subscription.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
			e, err := cloudbuildevent.ParseMessage(msg.Data)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: could not decode message data: %#v", msg)
				msg.Ack()
				return
			}

			fmt.Fprintf(os.Stdout, "Processing event: %#v\n", e)

			mu.Lock()
			defer mu.Unlock()

			if err := f.processGCB(ctx, e); err != nil {
				fmt.Fprintf(os.Stderr, "Error: cloud not process event: %s\n", err)

				msg.Ack()
				return
			}
			msg.Ack()
		})

		if err != nil {
			log.Fatal(err)
		}
	}
	errCh <- nil
}

func (f *Flow) Stop(ctx context.Context) {
	killCh <- true
}
