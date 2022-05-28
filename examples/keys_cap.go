package main

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/alextanhongpin/dataloader2"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func randSleep() {
	n := rand.Intn(5_000)
	duration := time.Duration(n) * time.Millisecond
	time.Sleep(duration)
}

func batchFetchNumbers(ctx context.Context, keys []int) (map[int]string, error) {
	fmt.Println("fetching keys", len(keys), keys)
	time.Sleep(5 * time.Second)

	result := make(map[int]string, len(keys))
	for _, k := range keys {
		result[k] = fmt.Sprint(k)
	}

	return result, nil
}

func main() {
	start := time.Now()
	defer func() {
		fmt.Println(time.Since(start))
	}()

	dl2, flush := dataloader2.New(
		context.Background(),
		batchFetchNumbers,
		dataloader2.WithBatchDuration[int, string](16*time.Millisecond),
		dataloader2.WithBatchCap[int, string](100),
		dataloader2.WithWorker[int, string](10),
	)
	defer flush()

	n := 1_000
	var wg sync.WaitGroup
	wg.Add(n)

	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()

			fmt.Println("start worker", i)
			randSleep()

			res, err := dl2.Load(i)
			if err != nil {
				panic(err)
			}
			_ = res
			fmt.Println("end worker", i, res)
		}(i)
	}

	wg.Wait()
}
