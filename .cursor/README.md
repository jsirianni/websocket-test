# Cursor Rules

This directory contains AI assistant rules for the WebSocket Test project. These rules ensure consistent code quality, architecture, and documentation across all AI-assisted development.

## Rule Files

### `go-websocket-project.mdc`
Project-specific standards for:
- Project structure (cmd/, server/, client/ organization)
- Server and client requirements
- Docker and container standards
- Build system (GoReleaser, Makefile)
- CI/CD workflows
- Kubernetes manifests
- Documentation structure

### `go-style.mdc`
Go language best practices:
- Naming conventions
- Function design
- Error handling
- Concurrency patterns
- Testing standards
- Import organization

### `documentation.mdc`
Documentation guidelines:
- File structure and organization
- README requirements
- Component documentation format
- Code examples and formatting
- Maintenance procedures

## How It Works

Cursor automatically applies these rules when:
- Working on matching files (based on glob patterns)
- Providing code suggestions
- Generating new code
- Answering questions about the project

## Modifying Rules

1. Edit the `.mdc` files directly
2. Use YAML frontmatter to control when rules apply:
   - `globs`: File patterns (e.g., `["**/*.go"]`)
   - `alwaysApply`: Whether to always use this rule
   - `description`: Brief description of the rule

3. Rules are written in Markdown for readability

## Benefits

- **Consistency**: Ensures all AI-generated code follows project standards
- **Onboarding**: New contributors get instant project context
- **Quality**: Enforces best practices automatically
- **Documentation**: Living documentation that guides development

## Learn More

- [Cursor Rules Documentation](https://docs.cursor.com/en/context/rules)

