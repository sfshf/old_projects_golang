package service_test

import (
	"context"
	"log"
	"sync"
	"testing"
	"time"

	"github.com/nextsurfer/doom-go/internal/service"
)

func TestStatisticMongoCall(t *testing.T) {
	ctx := context.WithValue(context.Background(), service.PerformanceInfoKey, &service.PerformanceInfo{
		StartedAt: time.Now(),
	})
	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			ts := time.Now()
			service.StatisticMongoCall(ctx, ts)
			wg.Done()
		}()
	}
	wg.Wait()
	performanceInfo := ctx.Value(service.PerformanceInfoKey).(*service.PerformanceInfo)
	if performanceInfo != nil {
		log.Println(time.Since(performanceInfo.StartedAt))
		log.Println(performanceInfo.MongoCallCount)
		log.Println(performanceInfo.MongoCallDuration)
	}
}
