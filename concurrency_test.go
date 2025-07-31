package common

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"testing"
	"time"
)

func TestConcurrentJob(t *testing.T) {
	type Args struct {
		Foo string
	}
	type Ret struct {
		Bar string
	}

	job := NewConcurrentJob(100, func(ctx context.Context, args Args) Ret {
		select {
		case <-ctx.Done():
			return Ret{Bar: "cancelled"}
		default:
			time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
		}

		return Ret{Bar: args.Foo}
	}, 5*time.Second)

	job.Start()

	go func() {
		for _, input := range []Args{
			{Foo: "1"},
			{Foo: "2"},
			{Foo: "3"},
			{Foo: "4"},
			{Foo: "5"},
			{Foo: "6"},
		} {
			job.Enqueue(input)
		}
		job.Stop()
	}()

	for res := range job.Results() {
		if res.Error != nil {
			log.Println("Error:", res.Error)
		} else {
			log.Println("Success:", res.Value)
		}
	}
}

func TestConcurrentJob_StopEarly(t *testing.T) {
	type Args struct{ Foo string }
	type Ret struct{ Bar string }

	job := NewConcurrentJob(3, func(ctx context.Context, args Args) Ret {
		select {
		case <-ctx.Done():
			return Ret{Bar: "cancelled"}
		default:
			time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
		}

		return Ret{Bar: args.Foo}
	}, 2*time.Second)

	job.Start()

	go func() {
		for i := 0; i < 20; i++ {
			job.Enqueue(Args{Foo: fmt.Sprintf("%d", i)})
			if i == 2 { // stop early after few items
				job.Stop()
				return
			}
		}
	}()

	for res := range job.Results() {
		if res.Error != nil {
			log.Println("Error:", res.Error)
		} else {
			log.Println("Success:", res.Value)
		}
	}
}
