package host

import (
	"github.com/NebulousLabs/go-upnp"
	"strconv"

	"github.com/NebulousLabs/Sia/build"
)

// managedForwardPort adds a port mapping to the router.
func (h *Host) managedForwardPort() error {
	// If the port is invalid, there is no need to perform any of the other
	// tasks.
	h.mu.RLock()
	port := h.port
	h.mu.RUnlock()
	portInt, err := strconv.Atoi(port)
	if err != nil {
		return err
	}
	if build.Release == "testing" {
		return nil
	}

	d, err := upnp.Discover()
	if err != nil {
		return err
	}
	err = d.Forward(uint16(portInt), "Sia Host")
	if err != nil {
		return err
	}

	h.log.Println("INFO: successfully forwarded port", port)
	return nil
}

// managedClearPort removes a port mapping from the router.
func (h *Host) managedClearPort() error {
	// If the port is invalid, there is no need to perform any of the other
	// tasks.
	h.mu.RLock()
	port := h.port
	h.mu.RUnlock()
	portInt, err := strconv.Atoi(port)
	if err != nil {
		return err
	}
	if build.Release == "testing" {
		return nil
	}

	d, err := upnp.Discover()
	if err != nil {
		return err
	}
	err = d.Clear(uint16(portInt))
	if err != nil {
		return err
	}

	h.log.Println("INFO: successfully unforwarded port", port)
	return nil
}
