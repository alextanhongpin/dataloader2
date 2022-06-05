package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/alextanhongpin/dataloader2"
)

func batchFetchNumbers(ctx context.Context, keys []int) (map[int]string, error) {
	fmt.Println("fetching keys", keys)
	time.Sleep(1 * time.Second)

	return nil, errors.New("database query error")
}

func main() {
	defer func(start time.Time) {
		fmt.Println(time.Since(start))
	}(time.Now())

	dl2, flush := dataloader2.New(context.Background(), batchFetchNumbers)
	defer flush()

	res, err := dl2.Load(1)
	fmt.Println("load(1):", res, err)

	result, err := dl2.LoadMany([]int{1, 2, 3, 4, 5, 4, 3, 2, 1})
	fmt.Println("loadMany:", result, err)

	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()

		res, err := dl2.Load(1)
		fmt.Println("go load(1):", res, err)
	}()

	go func() {
		defer wg.Done()

		fmt.Println("go primed:", dl2.Prime(1, "hello world"))

		res, err := dl2.Load(1)
		fmt.Println("go primed > load(1):", res, err)
	}()

	go func() {
		defer wg.Done()

		result, err := dl2.LoadMany([]int{1, 2, 3, 4, 5, 4, 3, 2, 1})
		fmt.Println("go loadMany:", result, err)
	}()

	wg.Wait()
}
