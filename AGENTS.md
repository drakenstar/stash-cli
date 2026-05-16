# Repo Notes

## Commands

- Use `go test ./...` for full verification; there is no Makefile, CI config, or
  separate lint/typecheck task in the repo.
- Run a focused test with `go test ./app -run TestName`, or swap `./app` for the
  package being changed (`./stash`, `./command`, `./config`, `./ui`).
- Use `go build ./...` to catch compile errors without running tests.

## Structure

- `main.go` wires config, the Stash backend, and the Bubble Tea TUI.
- `app/` owns the top-level Bubble Tea model plus scenes/galleries tabs and
  command-to-message handling.
- `command/` is the shell-like command parser/binder used by TUI commands; its
  struct tag key is `command`.
- `config/` loads `~/.stash-cli.json` first, then CLI flags override or append
  values.
- `stash/` contains the `Stash` interface, GraphQL client implementation, and a
  limited `file://` local backend.
- `ui/` contains reusable Bubble Tea widgets only.

## Runtime Gotchas

- CLI usage is `stash-cli [STASH INSTANCE]`; flags include `--stashInstance`,
  `--pathMapping`, `--openCommandURL`, `--openCommandScene`, `--openCommandGallery`,
  and `--debug`.
- A `file://` stash instance uses `stash.NewLocalStash` to browse local media,
  but delete/update/tag/studio/performer methods still panic as unimplemented.
- GraphQL queries are expressed with `github.com/hasura/go-graphql-client` struct
  tags; tests in `stash/` use `stash/schema.graphql` as the validation schema.

## Commit Defaults

- Subject line: lower case only, under 50 characters.
- Body: include a short blurb explaining the change; wrap lines at 80 characters.
