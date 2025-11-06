# AGENTS.md - Agent Guidelines for power-dra-driver

This document provides guidelines and useful commands for AI agents contributing to the power-dra-driver repository.

> **⚠️ IMPORTANT**: When making changes to Makefile targets, PR requirements, code generation workflows, verification steps, or any other information referenced in this document, **AGENTS.md must be updated accordingly** to keep it synchronized with the actual project state.

## Key Requirements for Contributors

### Legal Requirements

- **CLA Required**: All contributors MUST align with LICENSE

### Pull Request Labels

All code PRs MUST be labeled with one of:
- ⚠️ `:warning:` - major or breaking changes
- ✨ `:sparkles:` - feature additions
- 🐛 `:bug:` - patch and bugfixes
- 📖 `:book:` - documentation or proposals
- 🌱 `:seedling:` - minor or other

## Essential Make Targets

### Building

```bash
# Build binaries
make build
```

## Important Development Patterns

### Adding New Resources

all code in pkg or cmd directories
no custom code in vendor directory

### Testing Strategy

1. **Unit Tests**: Test individual functions/methods

## Pre-Submit Checklist for Agents

Before submitting a PR, ensure:

1. **Code is compiling and up to date**:
   ```bash
   make build
   ```

## Common Workflows

### Making Code Changes

1. Make your code changes
2. Verify everything: `make build`
3. Commit changes with descriptive message

### Updating Dependencies

1. Update `go.mod` or `hack/tools/go.mod` as needed

## Quick Reference

| Task | Command |
|------|---------|
| Ensure Build is successful | `make build` |
