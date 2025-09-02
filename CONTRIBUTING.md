# Contributing to go-qs

Thank you for your interest in contributing to **go-qs**! We welcome pull requests and issues to help improve this library.

## How to Contribute

### Pull Requests
- Fork the repository and create your branch from `main`.
- Write clear, concise commit messages.
- Add tests for new features or bug fixes.
- Ensure your code passes all tests (`go test`).
- Document your changes in the README if needed.
- Reference related issues in your PR description.
- Example: If you add support for a new query format, include a sample query and expected output.
- If your change is inspired by or compared to the JavaScript `qs` library, mention the relevant behavior and differences.

### Issues
- Search for existing issues before opening a new one.
- When reporting a bug, include:
  - Example query string
  - Expected output (in Go and, if possible, in JavaScript `qs`)
  - Actual output
  - Environment details (Go version, OS, etc.)
- For feature requests, describe the use case and, if relevant, how it works in the JavaScript `qs` library.

#### Example Issue Template
```
**Query:**
user[name]=Alice&user[age]=30

**Raised (go-qs):**
map[user:map[age:30 name:Alice]]

**Expected (JavaScript qs):**
{ user: { name: 'Alice', age: '30' } }
...

**Environment:**
goqs module version: v0.1.0
Go version: go1.21.0
OS: macOS
```

## Code Style
- Follow Go conventions and idioms.
- Use `gofmt` before submitting.

## Questions
If you have questions, open an issue or start a discussion.

---

Thank you for helping make go-qs better!
