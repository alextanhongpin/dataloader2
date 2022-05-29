package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/alextanhongpin/dataloader2"
)

func batchFetchNumbers(ctx context.Context, keys []int) (map[int]string, error) {
	fmt.Println("fetching keys", keys)
	time.Sleep(1 * time.Second)

	return nil, errors.New("database query error")
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
		fmt.Println(err)
	}
	fmt.Println(res)

	result, err := dl2.LoadMany([]int{1, 2, 3, 4, 5, 4, 3, 2, 1})
	if err != nil {
		panic(err)
	}

	for _, res := range result {
		fmt.Println(res.Unwrap())
	}

	go func() {

		res, err := dl2.Load(1)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(res)
	}()
	go func() {
		fmt.Println("primed", dl2.Prime(1, "hello world"))

		res, err := dl2.Load(1)
		if err != nil {
			fmt.Println("error", err)
		}
		fmt.Println("primed result", res)
	}()

	select {
	case <-time.After(2 * time.Second):
		fmt.Println("exiting")
		return
	}
}
