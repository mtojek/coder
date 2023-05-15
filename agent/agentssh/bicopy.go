package agentssh

import (
	"context"
	"io"
	"log"
	"sync"
)

// Bicopy copies all of the data between the two connections and will close them
// after one or both of them are done writing. If the context is canceled, both
// of the connections will be closed.
func Bicopy(ctx context.Context, c1, c2 io.ReadWriteCloser) {
	log.Println("bicopy is starting")

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	defer func() {
		err1 := c1.Close()
		log.Printf("c1.Close err: %s", err1)
		err2 := c2.Close()
		log.Printf("c2.Close err: %s", err2)
	}()

	var wg sync.WaitGroup
	copyFunc := func(dst io.WriteCloser, src io.Reader) {
		defer func() {
			wg.Done()
			// If one side of the copy fails, ensure the other one exits as
			// well.
			cancel()
		}()
		_, err := io.Copy(dst, src)
		log.Printf("io.Copy err: %s", err)
	}

	wg.Add(2)
	go copyFunc(c1, c2)
	go copyFunc(c2, c1)

	// Convert waitgroup to a channel so we can also wait on the context.
	done := make(chan struct{})
	go func() {
		defer close(done)
		wg.Wait()
	}()

	select {
	case <-ctx.Done():
	case <-done:
	}
}
