package servers

import (
	"errors"
	"net"
	"net/http"
	"sync"

	"github.com/aegis-run/aegis/pkg/telemetry/logger"
)

type base struct {
	mu          sync.Mutex
	isListening bool

	name string
}

func (b *base) listen(ln net.Listener) (ok bool) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.isListening {
		logger.Warn("server.already_listening", "srv", b.name)
		return false
	}

	b.isListening = true
	logger.Info("server.listening",
		"srv", b.name,
		"addr", ln.Addr().String(),
	)

	return true
}

func (b *base) isServeErr(err error) error {
	if err == nil || errors.Is(err, http.ErrServerClosed) || errors.Is(err, net.ErrClosed) {
		return nil
	}

	logger.Error("server.listen_failed",
		"srv", b.name,
		"error", err,
	)
	return err
}

func (b *base) shutdown() {
	logger.Info("server.shutdown_requested", "srv", b.name)
	b.mu.Lock()
	b.isListening = false
	b.mu.Unlock()
}
