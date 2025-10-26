package closer

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"hr-helper/internal/pkg/houston/loggy"
)

var (
	instance *closer
)

func init() {
	instance = newCloser(syscall.SIGTERM, syscall.SIGINT)
}

func SetShutdownTimeout(t time.Duration) {
	instance.shutdownTimeout = t
}

func Add(f ...func() error) {
	instance.add(f...)
}

func AddNoErr(f ...func()) {
	for _, fi := range f {
		instance.add(func() error {
			fi()
			return nil
		})
	}
}

func Wait() {
	instance.wait()
}

type closer struct {
	mu              sync.Mutex
	once            sync.Once
	done            chan struct{}
	funcs           []func() error
	shutdownTimeout time.Duration
}

func newCloser(sigs ...os.Signal) *closer {
	c := &closer{
		done:            make(chan struct{}),
		shutdownTimeout: time.Minute,
	}
	if len(sigs) > 0 {
		go func() {
			ch := make(chan os.Signal, 1)
			signal.Notify(ch, sigs...)
			<-ch
			signal.Stop(ch)
			c.closeAll()
		}()
	}

	return c
}

func (c *closer) add(f ...func() error) {
	c.mu.Lock()
	c.funcs = append(c.funcs, f...)
	c.mu.Unlock()
}

func (c *closer) wait() {
	<-c.done
}

func (c *closer) closeAll() {
	loggy.Warnln("received signal for graceful shutdown")

	c.once.Do(func() {
		defer close(c.done)

		c.mu.Lock()
		funcs := c.funcs
		c.funcs = nil
		c.mu.Unlock()

		errs := make(chan error, len(funcs))
		for _, f := range funcs {
			go func(f func() error) {
				errs <- f()
			}(f)
		}

		waitStopCh := make(chan struct{})
		go func() {
			defer close(waitStopCh)
			for i := 0; i < cap(errs); i++ {
				if err := <-errs; err != nil {
					loggy.Errorf("error returned from Closer: %v", err)
				}
			}
		}()

		t := time.NewTimer(c.shutdownTimeout)
		defer t.Stop()

		select {
		case <-t.C:
			loggy.Warnln("shutdown timed out")
		case <-waitStopCh:
		}
	})
}
