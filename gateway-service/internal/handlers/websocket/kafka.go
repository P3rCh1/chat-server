package websocket

import (
	"context"
	"errors"
	"sync/atomic"

	"github.com/segmentio/kafka-go"
)

var (
	ctxStopWorkers, stopWorkers = context.WithCancel(context.Background())
	fatalErrorLoggedFlag        atomic.Bool
)

func StartKafkaWorkers(
	ws *WS,
	n int,
) {
	for range n {
		go KafkaWorker(ws)
	}
}

func KafkaWorker(ws *WS) {
	for {
		select {
		case <-ctxStopWorkers.Done():
			return
		default:
			msg, err := ws.services.Kafka.Read()
			if err != nil {
				if errors.Is(err, kafka.ErrGenerationEnded) || errors.Is(err, kafka.ErrGroupClosed) {
					if fatalErrorLoggedFlag.CompareAndSwap(false, true) {
						stopWorkers()
						ws.services.Log.Error(
							"kafka worker fatal error",
							"error", err,
						)
					}
					return
				}
				if !errors.Is(err, context.DeadlineExceeded) {
					ws.services.Log.Error(
						"kafka worker read",
						"error", err,
					)
				}
				continue
			}
			broadcast(ws, msg)
		}
	}
}
