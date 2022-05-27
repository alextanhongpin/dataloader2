package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/alextanhongpin/dataloader2"
)

func randSleep() {
	duration := time.Duration(rand.Intn(1_000)) * time.Millisecond
	time.Sleep(duration)
}

func batchFetchNumbers(keys []int) (map[int]string, error) {
	fmt.Println("fetching keys", len(keys), keys)
	randSleep()

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
		batchFetchNumbers,
		dataloader2.WithBatchDuration[int, string](32*time.Millisecond),
		dataloader2.WithBatchCap[int, string](100),
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

			res, err := dl2.Load(rand.Intn(n))
			if err != nil {
				panic(err)
			}
			fmt.Println("end worker", i, res)
		}(i)
	}

	wg.Wait()
}
