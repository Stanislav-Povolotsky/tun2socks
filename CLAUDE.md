# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository

tun2socks is a userspace networking engine written in Go: it exposes a TUN device (or a raw
fd, or WinTun on Windows), routes all IP traffic captured from it through a userspace TCP/IP
stack (gVisor's `gvisor.dev/gvisor/pkg/tcpip`), and forwards each TCP/UDP flow out through a
configurable outbound proxy (SOCKS4/5, HTTP, Shadowsocks, SSH, relay, or direct).

This is a fork of `github.com/xjasonlyu/tun2socks`, rebranded as
`github.com/Stanislav-Povolotsky/tun2socks` (module path, README, Docker images/labels, CI
badges, issue-template links, security/code-of-conduct contacts) and diverged from upstream at
commit `dda1b10` with fork-only features — see "Fork-specific features" below. Attribution to
the original project is kept in the README credits section; don't "fix" it back.

## Commands

- Build the CLI for the current platform: `make tun2socks` (output: `build/tun2socks`)
- Build every release target (full OS/arch matrix + zips + Android AAR): mirrors
  `.github/workflows/release.yml` — `make -j releases` for the CLI zips, plus a separate
  `gomobile bind` step for the AAR (see "Release pipeline" below); not reproducible locally
  without the Android NDK/JDK
- Build one specific cross-compiled target: `make <os>-<arch>`, e.g. `make linux-arm64`,
  `make windows-amd64-v3`, `make darwin-arm64` (full list: `UNIX_ARCH_LIST`/
  `WINDOWS_ARCH_LIST` in `Makefile`)
- Run all tests: `go test ./...`
- Run one package's tests: `go test ./dns/...`; run one test: `go test ./dns/ -run TestIsFakeDNSQuery_TunnelHijack`
- Vet: `go vet ./...`
- Lint (matches CI): `golangci-lint run ./...` — CI (`make lint`) runs it once per
  `GOOS` (darwin/windows/linux/freebsd/openbsd) since some files are platform-gated
- Format: `gofumpt -w .` (import ordering enforced by `gci`, prefix group
  `github.com/Stanislav-Povolotsky/tun2socks`, per `.golangci.yaml`)

## Architecture

### Data flow

`core/device` (TUN device / fd / WinTun) → `core.CreateStack` (gVisor netstack setup:
`core/stack.go`, `core/tcp.go`, `core/udp.go`) → `tunnel.Tunnel` (`tunnel/tcp.go`
`handleTCPConn`, `tunnel/udp.go` `handleUDPConn`) → outbound `proxy.Proxy.DialContext`/
`DialUDP`.

Each accepted TCP/UDP flow gets a `metadata.Metadata` (src/dst as `netip.Addr`+port) built
from the gVisor endpoint's `TransportEndpointID`. Before dialing out, `tunnel/tcp.go`/
`tunnel/udp.go` check, in order:

1. `dns.IsFakeDNSQuery` — addressed to the configured fake-DNS listen address? Answer locally
   via `dns.HandleQuery` (fake IP pool), never touching the proxy.
2. `dns.IsHijackQuery` — any flow to port 53 with hijacking enabled? Redirect to the
   configured upstream resolver via `dialer.DialContext`/`ListenPacket`, bypassing the proxy.
3. Otherwise `dns.ProcessMetadata` (rewrites fake-IP destinations back to the real hostname
   when the proxy can resolve it itself), then a normal proxy dial.

### `engine` package — the public API

`engine.Key` (`engine/key.go`) is the single config struct: populated from CLI flags
(`main.go`) or a YAML file for the CLI, or set directly by an embedder — `engine.Key`/
`engine.Start`/`engine.Stop` are also the surface exposed by the Android AAR (`gomobile bind`
targets this package directly). `engine/engine.go`'s `start()`/`netstack()` wires `Key` into:
dialer socket options, the optional REST API, fake-DNS pool creation, DNS-hijack validation,
gVisor stack options (including TCP keepalive, see `core/option`), and finally
`core.CreateStack`. `engine/register.go` blank-imports every `proxy/*` package so they
self-register; `engine/parse.go` (+ `_unix`/`_windows` variants) parses the `-device`/`-proxy`
flags into concrete device/proxy instances.

### Proxy protocols

`proxy/proxy.go` defines the `Proxy` interface (`DialContext`, `DialUDP`); `proxy/protocol.go`
is a self-registering registry (`RegisterProtocol`/`Parse`, keyed by URL scheme) — each
`proxy/<name>` package (`direct`, `http`, `reject`, `relay`, `shadowsocks`, `socks4`,
`socks5`, `ssh`) registers itself in an `init()`. Wire-level codecs live under `transport/`
(`transport/socks4`, `transport/socks5`, `transport/shadowsocks`, `transport/simple-obfs`) and
are shared between the proxy implementations and anything else that needs to speak these
protocols directly.

### Fork-specific features (not in upstream)

- **Fake DNS** (`dns/fakedns.go`, `dns/server.go`, `component/fakeip`): answers
  configured-address DNS queries with fake IPs from a pool, remembering the real hostname so
  the proxy can resolve it later (`dns.ProcessMetadata`/pool `LookBack`). Originally answered
  only via a real OS-level UDP socket (`dns.ReCreateServer`), which doesn't work when a VPN
  implementation (e.g. Android `VpnService`) captures *all* traffic including DNS into the TUN
  device — so it's also answered natively in-tunnel, via `tunnel/tcp.go`'s
  `handleFakeDNSTCP`/`tunnel/udp.go`'s `handleFakeDNSUDP`, independent of whether the real
  socket bind succeeded.
- **DNS hijacking** (`dns/hijack.go`): simpler and complementary — redirects *any* flow to
  port 53 (any destination) to one trusted upstream resolver, for apps that hardcode their own
  DNS server. Checked after fake DNS in the tunnel dispatch so the two don't fight over the
  same listen address.
- **Configurable TCP keepalive**: `engine.Key.TCPKeepalive{IdleTime,Interval,Count}` (was
  hardcoded upstream); `core/option/option.go`'s `TCPSocketOption`s are applied per-endpoint in
  `core/tcp.go`. Useful behind a load balancer with a short inactive-flow timeout.

### Release pipeline (`.github/workflows/release.yml`)

On tag push, two jobs run: `release` (`ubuntu-latest`) does `make -j releases` to cross-compile
the full OS/arch matrix and zip each binary, then a `gomobile bind` step builds
`build/tun2socks-android.aar` from the `engine` package (JDK 17 + Android NDK
`27.2.12479018`), stripped via `-ldflags="-s -w" -trimpath`; everything under `build/*` is
uploaded as release assets via `softprops/action-gh-release`. `ios` (`macos-latest`, Xcode
preinstalled — Apple targets can only be built on macOS) does the equivalent for iOS:
`gomobile bind -target=ios` produces an XCFramework (device + simulator slices) at
`build/Tun2socks.xcframework`, zipped with `ditto` (the Apple-recommended way to archive a
framework bundle, preserves its internal symlinks) into `tun2socks-ios.xcframework.zip`, and
uploaded to the same release. Both jobs are this fork's own addition — the `engine` package
needed no source changes for either mobile target: `runtime.GOOS == "ios"` was already handled
in `engine/parse.go` (the 4-byte protocol-family header on the fd you get from
`NEPacketTunnelFlow` via the common private-KVC trick), and every platform-specific file
elsewhere already resolves correctly for `GOOS=ios` because Go treats `ios` as implied by both
the `unix` build tag and any `_darwin.go`-suffixed file (verified via `go list -f
'{{.GoFiles}}'` with `GOOS=ios`, since actually compiling for `ios` requires cgo + Xcode and
can't be checked on a non-macOS machine).

### Keeping in sync with upstream

Use the `sync-upstream` skill (`.claude/skills/sync-upstream/SKILL.md`, invoke with
`/sync-upstream <version>`) to pull changes from `xjasonlyu/tun2socks` — it diffs from this
fork's actual divergence point and skips/re-adapts rebranding reverts instead of doing a naive
merge.
