package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/alextanhongpin/dataloader2"
)

func batchFetchNumbers(keys []int) (map[int]string, error) {
	fmt.Println("fetching keys", keys)
	time.Sleep(1 * time.Second)

	return nil, errors.New("database query error")
}

func main() {
	start := time.Now()
	defer func() {
		fmt.Println(time.Since(start))
	}()

	dl2, flush := dataloader2.New(batchFetchNumbers)
	defer flush()

	res, err := dl2.Load(1)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(res)

	result := dl2.LoadMany([]int{1, 2, 3, 4, 5, 4, 3, 2, 1})
	for _, res := range result {
		fmt.Println(res.Unwrap())
	}
}
