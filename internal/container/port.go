package container

import (
	"fmt"
	"math/rand"
	"net"
	"time"
)

const (
	minPort = 10000
	maxPort = 65535
)

var reservedPorts = map[int]bool{
	22:   true, // SSH
	80:   true, // HTTP
	443:  true, // HTTPS
	8080: true, // Alternative HTTP
	// Add any ports that should not be used here
}

type portFinder struct {
	rng *rand.Rand
}

func newPortFinder() *portFinder {
	return &portFinder{
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}
func (pf *portFinder) findAvailablePort() (string, error) {
	for i := 0; i < 100; i++ { // Try up to 100 times
		port := pf.rng.Intn(maxPort-minPort+1) + minPort
		addr := fmt.Sprintf(":%d", port)

		if reservedPorts[port] {
			continue
		}

		listener, err := net.Listen("tcp", addr)
		if err == nil {
			listener.Close()
			return fmt.Sprintf("%d", port), nil
		}
	}

	return "", fmt.Errorf("could not find an available port after 100 attempts")
}
