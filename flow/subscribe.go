package flow

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"

	"cloud.google.com/go/pubsub"
)

var (
	countMu sync.Mutex
	killCh  chan bool
)

func (f *Flow) subscribe(ctx context.Context, errCh chan error) {
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
			var e Event

			if err := json.Unmarshal(msg.Data, &e); err != nil {
				fmt.Fprintf(os.Stderr, "Error: could not decode message data: %#v", msg)
				msg.Ack()
				return
			}

			fmt.Fprintf(os.Stdout, "Processing event: %#v.", e)

			countMu.Lock()
			if err := f.process(e); err != nil {
				fmt.Fprintf(os.Stderr, "Error: cloud not process event: %s", err)

				msg.Nack()
				return
			}
			msg.Ack()
			countMu.Unlock()
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
