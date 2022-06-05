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
	defer func(start time.Time) {
		fmt.Println(time.Since(start))
	}(time.Now())

	dl2, flush := dataloader2.New(context.Background(), batchFetchNumbers)
	defer flush()
	fmt.Println("primed:", dl2.Prime(1, "hello world"))

	res, err := dl2.Load(1)
	fmt.Println("load(1):", res, err)

	res, err = dl2.Load(2)
	fmt.Println("load(2):", res, err)
}
