package main

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/alextanhongpin/dataloader2"
)

func randSleep() {
	duration := time.Duration(rand.Intn(5_000)) * time.Millisecond
	time.Sleep(duration)
}

func batchFetchNumbers(ctx context.Context, keys []int) (map[int]string, error) {
	fmt.Println("fetching keys", len(keys), keys)
	randSleep()

	result := make(map[int]string, len(keys))
	for _, k := range keys {
		result[k] = fmt.Sprint(k)
	}

	return result, nil
}

func main() {
	defer func(start time.Time) {
		fmt.Println(time.Since(start))
	}(time.Now())

	dl2, flush := dataloader2.New(context.Background(), batchFetchNumbers)
	defer flush()

	n := 1_000
	var wg sync.WaitGroup
	wg.Add(n)

	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()

			fmt.Println("worker: start", i)
			randSleep()

			res, err := dl2.Load(rand.Intn(n))
			fmt.Println("worker: end", i, res, err)
		}(i)
	}

	wg.Wait()
}
