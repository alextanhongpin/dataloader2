package main

import (
	"context"
	"fmt"
	"time"

	"github.com/alextanhongpin/dataloader2"
)

func batchFetchNumbers(ctx context.Context, keys []int) (map[int]string, error) {
	fmt.Println("fetching keys", len(keys), keys)
	time.Sleep(1 * time.Second)

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

	dl2, flush := dataloader2.New(
		context.Background(),
		batchFetchNumbers,
		dataloader2.WithBatchCap[int, string](25),
		dataloader2.WithBatchWorker[int, string](4),
	)
	defer flush()

	n := 100

	numbers := make([]int, n)
	for i := 0; i < n; i++ {
		numbers[i] = i
	}

	result, err := dl2.LoadMany(numbers)
	fmt.Println("loadMany:", result, err)
}
