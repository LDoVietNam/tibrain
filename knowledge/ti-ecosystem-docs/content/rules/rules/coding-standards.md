# Coding Standards Rule

> **Priority**: P0
> **Location**: `Z:\10_WORKPLACE\Ti\content\rules\coding-standards.md`

---

## 🖊️ Coding Standards (BẮT BUỘC Follow)

### Go Standards
- **Formatter**: `gofmt` / `goimports` (không được dùng tabs khác)
- **Linter**: `golangci-lint` (config: `.golangci.yml`)
- **Naming**:
  - Package: lowercase, single word (e.g., `provider`, `router`)
  - Exported functions: PascalCase (e.g., `NewProvider`)
  - Unexported functions: camelCase (e.g., `validateConfig`)
  - Constants: PascalCase hoặc UPPER_SNAKE_CASE
- **Error Handling**: Trả về error (không panic), wrap errors với `fmt.Errorf("context: %w", err)`
- **Comments**: Godoc style cho exported functions
- **Testing**: File `*_test.go`, framework: built-in `testing`, assertion: `testify` nếu cần

**Check Command**:
```bash
golangci-lint run ./...
go test -cover ./...
```

---

### TypeScript/React Standards
- **Formatter**: Prettier (config: `.prettierrc`)
- **Linter**: ESLint (config: `.eslintrc.json`)
- **Naming**:
  - Components: PascalCase (e.g., `ProviderLogo.tsx`)
  - Functions/Hooks: camelCase (e.g., `useState`)
  - Constants: UPPER_SNAKE_CASE
  - Types/Interfaces: PascalCase (e.g., `ProviderLogoProps`)
- **Imports**: Absolute imports (từ `src/`), không relative (`../`)
- **Components**: Functional components + Hooks (không class components)
- **State**: useState, useReducer, hoặc Zustand/Redux (không Context API tràn lan)
- **Testing**: Vitest/Jest, React Testing Library

**Check Commands**:
```bash
npm run lint
npm test
npm run build
```

---

### Python Standards
- **Formatter**: Black (config: `pyproject.toml`)
- **Linter**: Pylint hoặc Flake8
- **Naming**:
  - Functions/variables: snake_case (e.g., `get_provider`)
  - Classes: PascalCase (e.g., `ProviderManager`)
  - Constants: UPPER_SNAKE_CASE
  - Private: `_leading_underscore` (e.g., `_validate_key`)
- **Type Hints**: Bắt buộc cho function signatures
- **Imports**: Stdlib → Third-party → Local (mỗi nhóm 1 dòng trống)
- **Testing**: Pytest

**Check Commands**:
```bash
black --check .
pylint src/
pytest --cov=src
```

---

### Rust Standards (nếu có)
- **Formatter**: `rustfmt`
- **Linter**: `clippy`
- **Naming**: snake_case cho functions/variables, PascalCase cho types
- **Error Handling**: `Result<T, E>`, `?` operator, không `unwrap()` (trừ tests)

**Check Commands**:
```bash
cargo fmt --check
cargo clippy
cargo test
```

---

## 📋 Pre-Commit Checklist (BẮT BUỘC)

- [ ] Code formatted (gofmt/Prettier/Black/rustfmt)
- [ ] Linting pass (golangci-lint/ESLint/Pylint/clippy)
- [ ] Tests pass (ít nhất 80% coverage cho code mới)
- [ ] No hardcoded secrets/API keys
- [ ] Comments cho exported functions/types (Godoc/JSDoc)
- [ ] No dead code (unused imports, variables, functions)

---

## 🔗 References
- **Go Best Practices**: https://go.dev/doc/effective_go
- **TypeScript Guidelines**: https://typescript.tv/best-practices/
- **Python PEP 8**: https://pep8.org/
- **Project `.editorconfig`**: (nếu có)
- **Review Rule**: `Z:\10_WORKPLACE\Ti\content\rules\core\review.md`
