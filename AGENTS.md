# AGENTS.md

Guidance for AI coding agents working in this repository. Read fully before making changes.

## Project overview

**Jalin v1.0.0** — High-performance desktop network utility app built with **MyGo** (`github.com/egoist/mygo` - Go backend + native GPU-drawn UI, no webview, pure Go without CGO).

Core features:
1. **Speed Test Benchmark**: measure download, upload, latency, jitter with real-time waveform chart streaming and custom velocity speedometer gauge.
2. **Live Network Monitor**: continuous real-time RX (download) and TX (upload) interface throughput monitoring via Windows kernel `GetIfTable2` with dual-channel sparkline graph and data transfer telemetry.
3. **History Storage**: persist results, settings, and proxy profiles in local **SQLite** (`modernc.org/sqlite`).
4. **Proxy Connect & Router**: connect through configured HTTP & SOCKS5 proxies with reachability testing and app-level traffic routing.
5. **Brand Identity**: geometric SVG logo (`assets/logo.svg`, `assets/logo-mark.svg`) following `logo-design-skill` principles.

UI & Layout:
- Left sidebar navigation (**Speed Test**, **Live Network**, **Test History**, **Proxy Router**, **Settings**).
- GPU-accelerated responsive native UI via `github.com/egoist/mygo/ui` and `chart`.
- Dark luxury cybernetic aesthetic, custom circular speedometer gauge, live waveform charts, tactile controls.

## Tech stack

| Layer | Choice |
|---|---|
| Runtime & Windowing | `github.com/egoist/mygo` (pure Go, zero cgo, native syscalls/purego) |
| UI Toolkit | `github.com/egoist/mygo/ui` (GPU rendered, flexbox layout, widgets, charts) |
| Live Network Monitor | `iphlpapi.dll` (`GetIfTable2`) Windows syscalls (pure Go, no CGO) |
| Database | SQLite via `modernc.org/sqlite` (pure Go, no CGO) |
| Proxy | `net/http` + `golang.org/x/net/proxy` (SOCKS5), HTTP proxy via `http.Transport.Proxy` |
| Speed Test | Multi-stream HTTP concurrent engine with live metrics callbacks |

## Structure

```
.
├── main.go                 # app bootstrap, window setup, main loop
├── assets/                 # official logo and vector assets
│   ├── logo.svg            # brand combination mark
│   └── logo-mark.svg       # transparent standalone icon
├── internal/
│   ├── netstat/            # real-time live network interface RX/TX monitor
│   ├── speedtest/          # download/upload/latency engine (no UI, no DB)
│   ├── store/              # SQLite access, migrations, CRUD
│   ├── proxy/              # proxy config, dialer, connection test
│   └── ui/                 # MyGo native UI views, gauge, charts, router
├── go.mod
└── AGENTS.md
```

If the template generated a different layout, follow the template and adapt; do not reorganize without being asked.

## Commands

```bash
go run .                                                          # run desktop application locally
go build -o jalin.exe .                                           # build Windows desktop binary (debug)
go build -trimpath -ldflags="-s -w -H=windowsgui" -o jalin.exe .  # build optimized lightweight binary (~16MB, no console)
go test ./...                                                     # run all unit tests
go vet ./...                                                      # run Go vet linter
gofmt -l .                                                        # verify formatting (must output nothing)
```

## Architecture rules

1. **Business logic lives in pure Go**, in `internal/*` (`speedtest`, `netstat`, `proxy`, `store`).
2. **UI layer in `internal/ui`**: Reactive state machine coordinated via `App` struct, GPU rendered through `github.com/egoist/mygo/ui`.
3. Keep logic unit-testable and free of UI dependencies.
4. Use `context.Context` for every long-running operation so it can be cleanly cancelled (speed tests, proxy probes).
5. UI state changes from background goroutines must be dispatched via `win.Update(...)` thread-safely.
6. **Code modularity & Line limits**:
   - Maintain a **250 LOC ceiling** per module/file whenever possible. Refactor into focused sub-components rather than oversized files.
   - Keep line lengths under 120 characters.
7. **Strict Iconography & No Emojis**:
   - **NEVER use emojis** in UI views, components, logs, or codebase.
   - Always use vector SVGs, native UI geometry drawing, or technical text badges (`SRC`, `TUN`, `DIR`, `DST`).

## Feature: speed test

- Measure: latency (min/avg), jitter, download Mbps, upload Mbps.
- Multi-stream parallel HTTP workers for download and upload with configurable phase duration and short warm-up discarded from result.
- Server endpoints are configurable in `internal/speedtest/servers.go`.
- Live sample callbacks during test: `{phase, mbps, latency, jitter, progress, bytesTotal}`.
- When proxy is enabled, traffic routes through configured proxy client.

## Feature: SQLite storage

- Driver: `modernc.org/sqlite` (pure Go, zero CGO).
- Tables: `speed_results`, `proxy_profiles`, `settings`, `daily_usage` (Internet Speed Meter daily traffic).
- Safe single-writer connection with WAL mode and foreign keys enabled.
- Parameterized queries throughout; timestamps handled via `parseDBTime`.

## Proxy support

- HTTP and SOCKS5 proxy support via `net/http` and `golang.org/x/net/proxy`.
- App-level traffic routing (speed tests and reachability probes).
- Live connectivity testing with latency feedback.

## UI and layout

- Pure Go native GPU-rendered instrument aesthetic with dark cybernetic theme.
- Top faceplate navigation: **Speed**, **Network**, **History**, **Proxy**, **Settings**.
- Bottom status strip: Real-time RX/TX throughput and test state.
- Dynamic responsive layout adapting to window dimensions (`wide(c)`).

## Boundaries

**Always**
- Run `go test ./...` and `go vet ./...` before finishing changes.
- Ensure `gofmt -l .` reports no unformatted files.
- Keep platform-specific Windows syscalls isolated in `netstat_windows.go` with `!windows` fallback.
- Strictly adhere to the No-Emoji policy: use icons/SVGs/badges instead.

**Never**
- Use emojis in any UI view, button, or label.
- Store proxy passwords in plain logs or export them in UI views.
- Introduce CGO dependencies.
- Make direct UI mutations from background goroutines without `win.Update`.