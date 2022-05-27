package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/alextanhongpin/dataloader2"
)

func batchFetchNumbers(keys []int) (map[int]string, error) {
	fmt.Println("fetching keys", keys)
	time.Sleep(1 * time.Second)

	result := make(map[int]string, len(keys))
	for _, k := range keys {
		if k < 4 {
			continue
		}
		result[k] = fmt.Sprint(k)
	}

	return result, nil
}

func main() {
	start := time.Now()
	defer func() {
		fmt.Println(time.Since(start))
	}()

	dl2, flush := dataloader2.New(batchFetchNumbers)
	defer flush()

	n := 10
	var wg sync.WaitGroup
	wg.Add(n)

	for i := 0; i < n; i++ {

		go func(i int) {
			defer wg.Done()

			fmt.Println("start worker", i)
			res, err := dl2.Load(i / 2)
			if err != nil {
				fmt.Println("worker error", i, err)
				return
			}
			fmt.Println("end worker", i, res)

		}(i)
	}

	wg.Wait()
}
