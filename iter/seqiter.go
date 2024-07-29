package iter

import (
	"iter"
	"sync"

	"github.com/sourcegraph/conc"
)

// SeqIterator can be used to configure the behaviour of ForEach. The zero value is
// safe to use with reasonable defaults.
//
// SeqIterator is also safe for reuse and concurrent use.
type SeqIterator[T any] SeqIterator2[T, struct{}]

// ForEach executes f in parallel over each element in input.
//
// ForEach always uses at most runtime.GOMAXPROCS goroutines.
// It takes roughly 2µs to start up the goroutines and adds
// an overhead of roughly 50ns per element of input. For
// a configurable goroutine limit, use a custom SeqIterator.
func (iter SeqIterator[T]) ForEach(input iter.Seq[T], f func(T) bool) {
	SeqIterator2[T, struct{}](iter).ForEach(
		func(yield func(T, struct{}) bool) {
			for v := range input {
				if !yield(v, struct{}{}) {
					return
				}
			}
		},
		func(v T, _ struct{}) bool { return f(v) },
	)
}

// ForEachSeq executes f in parallel over each element in input.
//
// ForEach always uses at most runtime.GOMAXPROCS goroutines.
// It takes roughly 2µs to start up the goroutines and adds
// an overhead of roughly 50ns per element of input. For
// a configurable goroutine limit, use a custom SeqIterator.
func ForEachSeq[T any](input iter.Seq[T], f func(T) bool) {
	SeqIterator[T]{}.ForEach(input, f)
}

// SeqIterator2 can be used to configure the behaviour of ForEach2. The zero value
// is safe to use with reasonable defaults.
//
// SeqIterator2 is also safe for reuse and concurrent use.
type SeqIterator2[T1, T2 any] struct {
	// MaxGoroutines controls the maximum number of goroutines
	// to use on this SeqIterator's methods.
	//
	// If unset, MaxGoroutines defaults to runtime.GOMAXPROCS(0).
	MaxGoroutines int
}

// ForEachSeq2 is the same as ForEach except it also provides the
// index of the element to the callback.
func ForEachSeq2[T1, T2 any](input iter.Seq2[T1, T2], f func(T1, T2) bool) {
	SeqIterator2[T1, T2]{}.ForEach(input, f)
}

// ForEach executes f in parallel over each element in input,
// using up to the SeqIterator's configured maximum number of
// goroutines.
//
// It is safe to mutate the input parameter, which makes it
// possible to map in place.
//
// It takes roughly 2µs to start up the goroutines and adds
// an overhead of roughly 50ns per element of input.
func (iter SeqIterator2[T1, T2]) ForEach(input iter.Seq2[T1, T2], f func(T1, T2) bool) {
	if iter.MaxGoroutines == 0 {
		// iter is a value receiver and is hence safe to mutate
		iter.MaxGoroutines = defaultMaxGoroutines()
	}

	type value struct {
		v1 T1
		v2 T2
	}

	sendCh := make(chan value, iter.MaxGoroutines)
	stopCh := make(chan struct{})
	stop := sync.OnceFunc(func() { close(stopCh) })

	go func() {
		defer close(sendCh)

		for v1, v2 := range input {
			select {
			case <-stopCh:
				return
			case sendCh <- value{v1, v2}:
			}
		}
	}()

	task := func() {
		defer stop()

		for v := range sendCh {
			if !f(v.v1, v.v2) {
				return
			}
		}
	}

	var wg conc.WaitGroup
	for i := 0; i < iter.MaxGoroutines; i++ {
		wg.Go(task)
	}
	wg.Wait()
}

// Slice returns a new sequence from a slice.
func Slice[T any](s []T) iter.Seq[*T] {
	return func(yield func(v *T) bool) {
		for i := range s {
			if !yield(&s[i]) {
				return
			}
		}
	}
}

// SliceIdx returns a new sequence of (index, value) from a slice.
func SliceIdx[T any](s []T) iter.Seq2[int, *T] {
	return func(yield func(i int, v *T) bool) {
		for i := range s {
			if !yield(i, &s[i]) {
				return
			}
		}
	}
}
