package main

import (
	"fmt"
	"time"

	"github.com/alextanhongpin/dataloader2"
)

func batchFetchNumbers(keys []int) (map[int]string, error) {
	fmt.Println("fetching keys", keys)
	time.Sleep(1 * time.Second)

	result := make(map[int]string, len(keys))
	for _, key := range keys {
		result[key] = fmt.Sprint(key)
	}
	fmt.Println("fetched keys")

	return result, nil
}

func main() {
	start := time.Now()
	defer func() {
		fmt.Println(time.Since(start))
	}()

	dl2, flush := dataloader2.New(batchFetchNumbers)
	// Immediately called flush.
	flush()

	primed := dl2.Prime(1, "hello world")
	fmt.Println("primed", primed)

	res, err := dl2.Load(1)
	if err != nil {
		fmt.Println("load error:", err)
	}
	fmt.Println("res:", res)

	res, err = dl2.Load(2)
	if err != nil {
		fmt.Println("load error:", err)
	}
	fmt.Println("res:", res)

	results, err := dl2.LoadMany([]int{1, 2, 3})
	if err != nil {
		fmt.Println("loadMany error:", err)
	}
	fmt.Println("results:", results)
	for _, res := range results {
		fmt.Println(res.Unwrap())
	}
}
