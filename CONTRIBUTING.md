# Zero to Secure L1 - Contributing Guidelines

Thank you for contributing to **Zero to Secure L1**! To maintain code quality and clean git history, please follow these guidelines when submitting code.

---

## 🌿 Branch Naming Convention

All feature branches and bug fixes must follow the pattern: `<type>/<short-description>`

Examples:
- `feat/avalanche-cli-launcher`
- `fix/deploy-output-parsing`
- `docs/update-architecture`
- `chore/add-ci-workflow`

Allowed types:
- `feat` - New features
- `fix` - Bug fixes
- `docs` - Documentation updates
- `chore` - Tooling, dependencies, or maintenance work
- `test` - Unit or integration test additions
- `refactor` - Code refactoring without changing functionality

---

## 📝 Commit Message Convention (Conventional Commits)

This repository enforces the [Conventional Commits](https://www.conventionalcommits.org/) specification using `commitlint` and `husky`.

### Format
`<type>(<optional-scope>): <short summary in present tense>`

Examples:
- `feat(cli): implement avalanche fuji deployment launcher`
- `fix(cli): resolve regex matching error for chain id`
- `docs: add commit rules and branch naming standards`
- `test(cli): add unit tests for deployment record parser`
- `chore: add github actions ci workflow`

### Allowed Commit Types
- `feat`: A new feature
- `fix`: A bug fix
- `docs`: Documentation only changes
- `chore`: Changes to build process, tooling, or auxiliary dependencies
- `refactor`: A code change that neither fixes a bug nor adds a feature
- `test`: Adding missing tests or correcting existing tests
- `perf`: A code change that improves performance
- `ci`: Changes to CI configuration files and scripts

> ⚠️ Commits that do not adhere to this format will be automatically rejected by the git `commit-msg` hook.

---

## 🧪 Local Testing & Verification

Before committing and pushing changes, verify your code across all module layers:

```bash
# 1. CLI (Go)
cd cli
go build ./...
go vet ./...
go test -v ./...

# 2. Dashboard Backend (Go)
cd ../dashboard/backend
go build ./...
go vet ./...

# 3. Dashboard Frontend (React + TS)
cd ../frontend
npm install
npm run build

# 4. Security Scanner (Python)
cd ../../security-scanner
pip install -r requirements.txt
```

