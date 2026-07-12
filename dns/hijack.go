package dns

import "sync"

var (
	hijackMu   sync.RWMutex
	hijackAddr string
)

// EnableHijack redirects any TCP/UDP flow to port 53 (any destination,
// unlike fake DNS which only answers queries to a specific configured
// address) to addr instead of dialing the proxy. Useful for forcing a
// trusted upstream resolver even for apps that hardcode their own DNS
// server, bypassing whatever the OS/network would otherwise use.
func EnableHijack(addr string) {
	hijackMu.Lock()
	hijackAddr = addr
	hijackMu.Unlock()
}

// DisableHijack turns DNS hijacking off.
func DisableHijack() {
	hijackMu.Lock()
	hijackAddr = ""
	hijackMu.Unlock()
}

// HijackTarget returns the configured upstream DNS server address, or ""
// if hijacking is disabled.
func HijackTarget() string {
	hijackMu.RLock()
	defer hijackMu.RUnlock()
	return hijackAddr
}

// IsHijackQuery reports whether a flow to dstPort should be redirected to
// the hijack target rather than dialed through the proxy.
func IsHijackQuery(dstPort uint16) bool {
	return dstPort == 53 && HijackTarget() != ""
}
