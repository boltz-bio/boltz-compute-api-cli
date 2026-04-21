# Boltz Compute CLI

The official CLI for the [Boltz Compute REST API](https://docs.boltz.bio/compute).

It is generated with [Stainless](https://www.stainless.com/).

<!-- x-release-please-start-version -->

## Installation

### Installing with Go

To test or install the CLI locally, you need [Go](https://go.dev/doc/install) version 1.25 or later installed.

```sh
go install 'github.com/boltz-bio/boltz-compute-api-cli/cmd/boltz-api@latest'
```

Once you have run `go install`, the binary is placed in your Go bin directory:

- **Default location**: `$HOME/go/bin` (or `$GOPATH/bin` if GOPATH is set)
- **Check your path**: Run `go env GOPATH` to see the base directory

If commands aren't found after installation, add the Go bin directory to your PATH:

```sh
# Add to your shell profile (.zshrc, .bashrc, etc.)
export PATH="$PATH:$(go env GOPATH)/bin"
```

<!-- x-release-please-end -->

### Running Locally

After cloning the git repository for this project, you can use the
`scripts/run` script to run the tool locally:

```sh
./scripts/run args...
```

## Usage

The CLI follows a resource-based command structure:

```sh
boltz-api [resource] <command> [flags...]
```

```sh
boltz-api predictions:structure-and-binding start \
  --api-key 'My API Key' \
  --input '{entities: [{chain_ids: [string], modifications: [{residue_index: 0, type: ccd, value: value}], type: protein, value: value}]}' \
  --model boltz-2.1
```

For details about specific commands, use the `--help` flag.

### Environment variables

| Environment variable    | Required | Default value |
| ----------------------- | -------- | ------------- |
| `BOLTZ_COMPUTE_API_KEY` | no       | `null`        |

OAuth mode can also be configured with:

- `BOLTZ_COMPUTE_AUTH_ISSUER_URL`
- `BOLTZ_COMPUTE_AUTH_CLIENT_ID`
- `BOLTZ_COMPUTE_AUTH_SCOPE` (comma-separated)
- `BOLTZ_COMPUTE_AUTH_AUDIENCE`
- `BOLTZ_COMPUTE_AUTH_AUTHORIZATION_URL`
- `BOLTZ_COMPUTE_AUTH_TOKEN_URL`
- `BOLTZ_COMPUTE_AUTH_USERINFO_URL`
- `BOLTZ_COMPUTE_AUTH_REVOCATION_URL`
- `BOLTZ_COMPUTE_ORG`
- `BOLTZ_COMPUTE_NO_BROWSER`
- `BOLTZ_COMPUTE_LISTEN_PORT`

### Global flags

- `--api-key` (can also be set with `BOLTZ_COMPUTE_API_KEY` env var)
- `--help` - Show command line usage
- `--debug` - Enable debug logging (includes HTTP request/response details)
- `--version`, `-v` - Show the CLI version
- `--base-url` - Use a custom API backend URL
- `--format` - Change the output format (`auto`, `explore`, `json`, `jsonl`, `pretty`, `raw`, `yaml`)
- `--format-error` - Change the output format for errors (`auto`, `explore`, `json`, `jsonl`, `pretty`, `raw`, `yaml`)
- `--transform` - Transform the data output using [GJSON syntax](https://github.com/tidwall/gjson/blob/master/SYNTAX.md). On paginated or streamed list commands, the transform runs on each item.
- `--transform-error` - Transform the error output using [GJSON syntax](https://github.com/tidwall/gjson/blob/master/SYNTAX.md)
- `--auth-issuer-url` - OIDC issuer URL used for OAuth login and bearer-token refresh
- `--auth-client-id` - OAuth client ID for public-client login
- `--auth-scope` - OAuth scope to request (repeatable)
- `--auth-audience` - Optional OAuth audience
- `--auth-authorization-url` - Override the discovered authorization endpoint
- `--auth-token-url` - Override the discovered token endpoint
- `--auth-userinfo-url` - Override the discovered userinfo endpoint
- `--auth-revocation-url` - Override the discovered revocation endpoint
- `--org` - Persist or override the local selected organization
- `--no-browser` - Print the OAuth URL without opening a browser
- `--listen-port` - Bind the OAuth loopback listener to a specific port

### OAuth authentication

The CLI supports API-key mode and OAuth bearer-token mode. When `--api-key` or
`BOLTZ_COMPUTE_API_KEY` is present, API-key mode wins. Otherwise the CLI uses a
stored OAuth session if one matches the configured issuer, client ID, audience,
and scopes.

Use API-key mode for CI and agent automation. The stored OAuth session is a
human login flow with local state, refresh, and browser/loopback behavior.

Start a login flow with:

```sh
boltz-api auth login \
  --auth-issuer-url 'https://issuer.example.com' \
  --auth-client-id 'public-client-id'
```

For remote or headless usage, print the URL instead of opening a browser:

```sh
boltz-api auth login \
  --no-browser \
  --auth-issuer-url 'https://issuer.example.com' \
  --auth-client-id 'public-client-id'
```

Available auth commands:

- `boltz-api auth login`
- `boltz-api auth logout`
- `boltz-api auth whoami`
- `boltz-api auth status`
- `boltz-api auth validate`
- `boltz-api auth switch-org <org>`

Command roles:

- `auth whoami` - concise local identity and current mode
- `auth status` - stable machine-readable auth diagnostics without refreshing
- `auth validate` - local auth check that may refresh an expired OAuth access token

`auth status` and `auth validate` return structured output. They exit with code
`1` when no usable auth mode is available. `auth status` remains read-only;
`auth validate` may refresh an expired OAuth access token using the stored
refresh token. In API-key mode, `auth validate` confirms that an API key is
configured locally; it does not make a server round-trip.

The CLI stores non-secret auth configuration in:

- `~/.config/boltz-compute/config.yaml`
- `~/.cache/boltz-compute/session.json`

Refresh tokens are stored in the OS keychain when available, with a fallback to:

- `~/.config/boltz-compute/credentials.json`

### Passing files as arguments

To inline file contents into request values, you can use the `@myfile.ext`
syntax:

```bash
boltz-api <command> --arg @abe.jpg
```

Files can also be passed inside JSON or YAML blobs:

```bash
boltz-api <command> --arg '{image: "@abe.jpg"}'
# Equivalent:
boltz-api <command> <<YAML
arg:
  image: "@abe.jpg"
YAML
```

To parse a file as structured JSON or YAML and inject the parsed object or
array, use `@json://...` or `@yaml://...`:

```bash
boltz-api predictions:structure-and-binding start \
  --input @json:///tmp/input.json

boltz-api predictions:structure-and-binding start <<'YAML'
input:
  entities: "@yaml:///tmp/entities.yaml"
model: boltz-2.1
YAML
```

If you need to pass a string literal that begins with an `@` sign, you can
escape the `@` sign to avoid accidentally passing a file.

```bash
boltz-api <command> --username '\@abe'
```

#### Explicit encoding

For JSON endpoints, the CLI tool does filetype sniffing to determine whether the
file contents should be sent as a string literal (for plain text files) or as a
base64-encoded string literal (for binary files). If you need to explicitly send
the file as either plain text or base64-encoded data, you can use
`@file://myfile.txt` (for string encoding) or `@data://myfile.dat` (for
base64-encoding). Use `@json://...` or `@yaml://...` only when you want the CLI
to parse the referenced file and inject structured data. Note that absolute
paths will begin with `@file://`, `@data://`, `@json://`, or `@yaml://`,
followed by a third `/` (for example, `@file:///tmp/file.txt`).

```bash
boltz-api <command> --arg @data://file.txt
```

### CLI flags vs piped JSON/YAML

Some array-valued request body fields are exposed as repeatable singular flags
for CLI ergonomics. Repeat the flag to add multiple entries:

```bash
boltz-api small-molecule:library-screen start \
  --molecule '{smiles: CCO}' \
  --molecule '{smiles: CCN}' \
  --target @json:///tmp/target.json

boltz-api protein:library-screen start \
  --protein @json:///tmp/protein-a.json \
  --protein @json:///tmp/protein-b.json \
  --target @json:///tmp/target.json
```

When piping JSON or YAML on stdin, the CLI merges that data onto the HTTP
request body, so you must use API body field names, not CLI flag names:

```bash
boltz-api small-molecule:library-screen start <<'YAML'
molecules:
  - smiles: CCO
  - smiles: CCN
target: {}
YAML

boltz-api protein:library-screen start <<'YAML'
proteins:
  - {}
  - {}
target: {}
YAML
```

Use `--help` on a specific command to see the repeatable flag names it accepts.

### Transform behavior

`--transform` applies to the whole response for single-object commands. On
paginated or streamed list commands, it applies to each emitted item.

Examples:

```bash
# Per-item extraction on list output
boltz-api small-molecule:library-screen list-results \
  --screen-id sm_scr_123 \
  --transform 'input_molecule.id'

# Whole-list reshaping or aggregation is better handled with jq
boltz-api small-molecule:library-screen list-results \
  --screen-id sm_scr_123 \
  --format raw | jq '.data[] | {id, score}'
```

Array-root expressions such as `#.{...}` are not the right tool in streamed
per-item mode.

## Linking different Go SDK versions

You can link the CLI against a different version of the Boltz Compute Go SDK
for development purposes using the `./scripts/link` script.

To link to a specific version from a repository (version can be a branch,
git tag, or commit hash):

```bash
./scripts/link github.com/org/repo@version
```

To link to a local copy of the SDK:

```bash
./scripts/link ../path/to/boltzcompute-go
```

If you run the link script without any arguments, it will default to `../boltzcompute-go`.
