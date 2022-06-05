package main

import (
	"context"
	"fmt"
	"time"

	"github.com/alextanhongpin/dataloader2"
)

func batchFetchNumbers(ctx context.Context, keys []int) (map[int]string, error) {
	fmt.Println("batchFetchNumbers: fetching", keys)
	select {
	case <-time.After(5 * time.Second):
		result := make(map[int]string, len(keys))
		for _, key := range keys {
			result[key] = fmt.Sprint(key)
		}
		fmt.Println("batchFetchNumbers: fetched")

		return result, nil
	case <-ctx.Done():
		fmt.Println("batchFetchNumbers: aborted")
		return nil, ctx.Err()
	}
}

func main() {
	defer func(start time.Time) {
		fmt.Println(time.Since(start))
	}(time.Now())
	done := make(chan bool)

	dl2, flush := dataloader2.New(context.Background(), batchFetchNumbers)

	fmt.Println("primed:", dl2.Prime(1, "hello world"))

	res, err := dl2.Load(1)
	fmt.Println("load(1):", res, err)

	go func() {
		time.Sleep(500 * time.Millisecond)
		fmt.Println("flushing")
		flush()
		close(done)
	}()

	res, err = dl2.Load(2)
	fmt.Println("load(2):", res, err)

	results, err := dl2.LoadMany([]int{1, 2, 3})
	fmt.Println("loadMany:", results, err)
	<-done
}
