# Markdown Linting and Formatting Tools

This document explains how to use markdown linting and formatting tools for this project.

## Option 1: Using markdownlint-cli (Recommended)

### Installation
```bash
# Install globally
npm install -g markdownlint-cli

# Or install locally in project
npm install --save-dev markdownlint-cli
```

### Usage
```bash
# Check all markdown files
markdownlint .

# Fix fixable issues automatically
markdownlint --fix .

# Check specific file
markdownlint README.md
```

### Configuration
The project includes `.markdownlint.yaml` with our formatting rules.

## Option 2: Using Prettier

### Installation
```bash
npm install --save-dev prettier
```

### Usage
```bash
# Format markdown files
npx prettier --write "**/*.md"

# Check formatting
npx prettier --check "**/*.md"
```

## Option 3: Manual Validation

Use the validation script:
```bash
./scripts/validate_markdown.sh
```

## Integration with CI/CD

Add to your workflow:
```yaml
- name: Lint Markdown
  run: |
    npm install -g markdownlint-cli
    markdownlint .
```