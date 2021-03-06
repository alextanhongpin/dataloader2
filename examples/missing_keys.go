package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/alextanhongpin/dataloader2"
)

func batchFetchNumbers(ctx context.Context, keys []int) (map[int]string, error) {
	fmt.Println("fetching keys", keys)
	time.Sleep(1 * time.Second)

	result := make(map[int]string, len(keys))
	for _, k := range keys {
		if k < 3 {
			continue
		}
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

	n := 10

	var wg sync.WaitGroup
	wg.Add(n)

	for i := 0; i < n; i++ {

		go func(i int) {
			defer wg.Done()

			fmt.Println("worker:", i, "key:", i/2)
			res, err := dl2.Load(i / 2)
			fmt.Println("worker:", i, res, err)
		}(i)
	}

	wg.Wait()
}
