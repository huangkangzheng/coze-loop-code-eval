# Repository Guidelines

## Project Structure & Module Organization
- `backend/`: Go services, Kitex/Hertz APIs, and domain modules under `modules/`; shared tooling lives in `pkg/` and generated stubs in `kitex_gen/`.
- `frontend/`: Rush-managed TypeScript workspace; UI app sits in `apps/cozeloop/`, while reusable packages, configs, and plugins reside under `packages/` and `infra/`.
- `common/`: shared scripts, Rush autoinstallers, and repository-level config used by both stacks.
- `idl/` and `release/`: IDL sources for API generation plus deployment assets (Docker Compose, Helm charts, images) and environment templates.

## Build, Test, and Development Commands
- `make compose-up`: launch the full stack via Docker Compose using configs in `release/deployment/docker-compose/`.
- `cd backend && go test -gcflags="all=-N -l" ./...`: run backend unit tests with debugging flags (preferred before any PR).
- `cd frontend && rush update`: install or refresh workspace dependencies with pnpm via Rush.
- `cd frontend/apps/cozeloop && rushx dev`: start the web UI at `http://localhost:8090`.
- `cd frontend/apps/cozeloop && rushx build` / `rushx test`: produce production bundles or execute Vitest suites.
- `rush update-api`: regenerate TypeScript API bindings from the latest IDL files.

## Coding Style & Naming Conventions
- Go code must remain `gofmt`/`goimports` clean; package and file names stay lowercase with underscores avoided, and exported identifiers include GoDoc-style comments.
- TypeScript follows the repository ESLint/Tailwind presets (`@coze-arch/*` configs). Use PascalCase for React components, camelCase for hooks and utilities, and kebab-case for file paths.
- Keep error handling consistent with existing `errorx` helpers and prefer dependency injection patterns already wired in `backend/infra`.

## Testing Guidelines
- Place Go tests next to implementation files as `*_test.go`, using table-driven cases and mocks generated via `mockgen` where needed.
- Frontend tests live alongside source files as `*.test.ts(x)` or `*.spec.ts(x)` and should cover state stores, hooks, and critical UI flows. Use `@vitest/coverage-v8` when validating coverage locally.
- Add regression tests when fixing bugs and ensure suites pass before requesting review.

## Commit & Pull Request Guidelines
- Commits follow Conventional Commits (`feat:`, `fix:`, `docs:` …); scope modules when it clarifies impact (e.g., `feat(prompt): add batch runner`).
- Rebase on `main`, keep commits logically grouped, and include short bodies or footers for breaking changes and issue references.
- Pull requests must describe intent, list validation steps (commands run), link related issues, and attach UI screenshots or logs for user-facing updates. Ensure CI passes and that reviewers can reproduce results with documented commands.

## Security & Configuration Tips
- Store secrets only in local `.env` files or environment variables; never commit credentials from `release/deployment/docker-compose/conf/model_config.yaml`.
- Before deploying, confirm cloud access keys and model endpoints are set per environment, and avoid exposing internal services when testing remote integrations.
