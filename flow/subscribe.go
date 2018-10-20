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
)

type event struct {
	repo string
}

func (f *Flow) subscribe() {
	ctx := context.Background()
	err := subscription.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		var e event

		if err := json.Unmarshal(msg.Data, &e); err != nil {
			fmt.Fprintf(os.Stderr, "Error: could not decode message data: %#v", msg)
			msg.Ack()
			return
		}

		fmt.Fprintf(os.Stdout, "Processing event: %#v.", e)

		countMu.Lock()
		if err := f.process(e); err != nil {
			fmt.Fprintf(os.Stdout, "Error: cloud not process event: %s", err)

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
