package dataloader2_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/alextanhongpin/dataloader2"
)

func TestLoad(t *testing.T) {
	t.Parallel()

	fetchNumbers := func(ctx context.Context, keys []int) (map[int]string, error) {
		fmt.Println("fetching keys", keys)

		select {
		case <-time.After(100 * time.Millisecond):
			result := make(map[int]string, len(keys))
			for _, key := range keys {
				result[key] = fmt.Sprint(key)
			}

			return result, nil
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	ctx := context.Background()
	dl2, flush := dataloader2.New(ctx, fetchNumbers)
	t.Cleanup(flush)

	t.Run("load", func(t *testing.T) {
		t.Parallel()

		res, err := dl2.Load(1)
		if err != nil {
			t.FailNow()
		}

		if exp, got := "1", res; exp != got {
			t.Fatalf("expected %v, got %v", exp, got)
		}
	})

	t.Run("loadMany", func(t *testing.T) {
		t.Parallel()

		keys := []int{100, 200, 300}
		res, err := dl2.LoadMany(keys)
		if err != nil {
			t.FailNow()
		}

		if exp, got := 3, len(res); exp != got {
			t.Fatalf("expected %v, got %v", exp, got)
		}
		for _, key := range keys {
			if exp, got := fmt.Sprint(key), res[key]; exp != got {
				t.Fatalf("expected %v, got %v", exp, got)
			}
		}
	})

	t.Run("prime", func(t *testing.T) {
		t.Parallel()

		if !dl2.Prime(42, "the meaning of life") {
			t.FailNow()
		}

		res, err := dl2.Load(42)
		if err != nil {
			t.FailNow()
		}

		if exp, got := "the meaning of life", res; exp != got {
			t.Fatalf("expected %v, got %v", exp, got)
		}
	})
}

func TestMissingKeys(t *testing.T) {
	t.Parallel()

	fetchNumbers := func(ctx context.Context, keys []int) (map[int]string, error) {
		return nil, nil
	}

	ctx := context.Background()
	dl2, flush := dataloader2.New(ctx, fetchNumbers)
	t.Cleanup(flush)

	t.Run("missing keys", func(t *testing.T) {
		t.Parallel()

		_, err := dl2.Load(1)
		if exp, got := true, errors.Is(err, dataloader2.ErrKeyNotFound); exp != got {
			t.Fatalf("expected %v, got %v", exp, got)
		}
	})

	t.Run("no keys", func(t *testing.T) {
		t.Parallel()

		_, err := dl2.LoadMany(nil)
		if exp, got := true, errors.Is(err, nil); exp != got {
			t.Fatalf("expected %v, got %v", exp, got)
		}
	})
}
