package main

import (
	"context"
	"fmt"
	"time"

	"github.com/alextanhongpin/dataloader2"
)

func batchFetchNumbers(ctx context.Context, keys []int) (map[int]string, error) {
	fmt.Println("fetching keys", keys)
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

	dl2, flush := dataloader2.New(context.Background(), batchFetchNumbers)
	defer flush()

	res, err := dl2.Load(1)
	if err != nil {
		panic(err)
	}
	fmt.Println(res)

	result, err := dl2.LoadMany([]int{1, 2, 3, 4, 5, 4, 3, 2, 1})
	if err != nil {
		panic(err)
	}

	for _, res := range result {
		fmt.Println(res.Unwrap())
	}
}
