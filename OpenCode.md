## 🛠️ Build/Lint/Test Commands

### 🚀 Build
```bash
GO111MODULE=on go build -mod=vendor
```

### ✅ Lint
```bash
golangci-lint run
```

### 🧪 Test (single test)
```bash
go test -v ./... -run ^TestNamePattern
```

## 📜 Code Style Guidelines
- **Imports**: Sort with `goimports`, group standard/libs first
- **Formatting**: `gofmt -s` + `goimports` (run pre-commit hook)
- **Types**: Prefer explicit types over interfaces where possible
- **Naming**: snake_case for variables, camelCase for types
- **Errors**: Always check returns, use `errors.New()`/`fmt.Errorf()`

## 📁 Configuration
- Cursor rules: `.cursor/rules/` (add your custom rules here)
- Copilot instructions: `.github/copilot-instructions.md` (document your preferences)

## 📌 Notes
- All CLI commands assume `GOPATH` is set
- Use `go mod tidy` after dependency changes
- Tests must have `t.Helper()` for cleaner output

(20 lines max - expand as needed)