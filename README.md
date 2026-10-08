# lazymihomo

lazymihomo is a terminal node switcher for [mihomo](https://github.com/MetaCubeX/mihomo) (Clash.Meta). It takes the interaction style of [lazygit](https://github.com/jesseduffield/lazygit) — full-screen lists, keyboard driven, a keybinding bar at the bottom — and narrows the feature set of [metacubexd](https://github.com/MetaCubeX/metacubexd)'s proxies page down to one job: **pick a node and test its latency, without leaving the terminal.**

It talks to a **running** mihomo instance through its external-controller REST API. It does not manage the core, edit config files, or spawn processes — bare metal, systemd, and Docker deployments all work the same.

```
╭─ Proxies ──────────────────────────────────────────────────────────────────╮
│GROUPS (1)                        NODES · 🚀 节点选择 (Selector) · 47        │
│❯ 🚀 节点选择 (Selector)         │❯ ● ♻️ 自动选择                       89ms │
│                                  │    🇭🇰 香港Y01                      120ms │
│                                  │    🇭🇰 香港Y02 | IEPL               355ms │
│                                  │    🇯🇵 日本Y01 | IEPL               89ms │
│                                  │    ...                                     │
│                                  │                                            │
│· switched → 香港Y01                                                          │
╰────────────────────────────────────────────────────────────────────────────╯
 j/k move • enter switch node • d test node • t test group • s sort • q quit
```

## Features

- **Two-column proxies view**: policy groups on the left, the selected group's nodes with latency on the right.
- **Manual control only**: the groups column lists `Selector` groups exclusively. URLTest/Fallback groups pick their node automatically and reject manual switching, so they — together with mihomo's synthetic `GLOBAL` — are hidden. An auto group still appears as a regular entry inside a selector if your config puts it there.
- **Latency tests**: probe one node (`d`) or the whole group (`t`); sort by delay (`s`, timeouts and untested nodes sink to the bottom). Colors: green < 200 ms, yellow < 500 ms, red ≥ 500 ms or timeout; `—` means never tested; `●` marks the live node.
- **Provider-aware lookups**: mihomo keeps subscription nodes out of its `/proxies` map, so lazymihomo merges `/proxies` with every provider from `/providers/proxies` and routes delay probes for provider nodes through the provider-scoped healthcheck endpoint.
- **No defaults**: the controller address is never assumed (see Configuration).

## Build & run

Requires Go ≥ 1.27.

```bash
make build          # produces bin/lazymihomo
./bin/lazymihomo
```

## Configuration

Resolution order, per field: `--url` / `--secret` flags > `LAZYMIHOMO_URL` / `LAZYMIHOMO_SECRET` environment variables > interactive dialog.

- With nothing configured, lazymihomo first shows a centered dialog asking for the controller URL and secret. The URL is required; the secret may be left empty. `tab` switches fields, `enter` confirms, `esc` cancels.
- To skip the dialog every launch, export the variables:

```bash
export LAZYMIHOMO_URL=http://127.0.0.1:9090
export LAZYMIHOMO_SECRET=your-external-controller-secret
```

`lazymihomo --check` verifies connectivity (hits `/version`, counts nodes and groups). It is non-interactive: without a configured URL it errors out immediately, which makes it script-friendly.

## Keybindings

| Key | Action |
|---|---|
| `j` `k` `g` `G` | move / jump top / bottom |
| `h` `l` `tab` | switch focus between the groups and nodes columns |
| `enter` | switch the group's selected node |
| `d` | latency-test the highlighted node |
| `t` | latency-test every node in the group |
| `s` | toggle sort-by-delay |
| `r` | refresh |
| `?` | help overlay |
| `q` / `ctrl+c` | quit |

## Development

```bash
make test                       # unit tests: API client (httptest), config, formatting, headless rendering
make vet
go run ./cmd/dumpview           # render one frame to stdout without a terminal (layout debugging)
go run ./cmd/probe              # run the API client against a live controller (connectivity debugging)
```

Project layout:

```
main.go                     flags, dialog/launch flow, --check
internal/
├── api/                    mihomo REST client (proxies, groups, providers, delay routes)
├── appconfig/              flag > env resolution, no defaults
├── format/                 byte / bandwidth / duration formatting
└── ui/
    ├── app.go              root Bubble Tea model: box layout, footer, help overlay
    ├── prompt.go           connection dialog (url + secret)
    ├── theme/              lipgloss styles, delay colors, bordered box
    ├── components/         scrollable cursor list with ANSI-aware truncation
    └── panels/             the proxies view (groups × nodes, switching, delay tests)
cmd/
├── dumpview/               one-frame renderer for layout debugging
└── probe/                  live API exerciser for connectivity debugging
```

Dependencies: [bubbletea](https://github.com/charmbracelet/bubbletea) · [bubbles](https://github.com/charmbracelet/bubbles) · [lipgloss](https.com/charmbracelet/lipgloss) · [x/ansi](https://github.com/charmbracelet/x/ansi).
