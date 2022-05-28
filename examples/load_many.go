package main

import (
	"context"
	"fmt"
	"runtime"
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
	start := time.Now()
	defer func() {
		fmt.Println(time.Since(start))
	}()

	fmt.Println("worker", runtime.NumCPU())

	dl2, flush := dataloader2.New(
		context.Background(),
		batchFetchNumbers,
		dataloader2.WithBatchCap[int, string](50),
	)
	defer flush()

	n := 1_00

	numbers := make([]int, n)
	for i := 0; i < n; i++ {
		numbers[i] = i
	}
	result, err := dl2.LoadMany(numbers)
	if err != nil {
		panic(err)
	}

	for _, res := range result {
		fmt.Println(res.Unwrap())
	}
}
