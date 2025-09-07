# Markdown Style Guide

This document provides guidelines for writing and formatting markdown files in the go-queue project.

## Line Breaks and Spacing

### Question: Do I always add `<br>` tag to change line?

**Answer: No, you don't always need `<br>` tags for line breaks in markdown files.**

Markdown provides several ways to create line breaks, and `<br>` tags are rarely necessary:

### 1. Paragraph Breaks (Recommended)

Use **double newlines** (blank line) to separate paragraphs:

```markdown
This is the first paragraph.

This is the second paragraph.
```

**Result:**
This is the first paragraph.

This is the second paragraph.

### 2. Line Breaks Within Paragraphs

Use **two trailing spaces** followed by a newline for line breaks within the same paragraph:

```markdown
First line  
Second line in same paragraph
```

**Result:**
First line  
Second line in same paragraph

### 3. HTML `<br>` Tags (Use Sparingly)

Only use `<br>` tags when you specifically need HTML rendering or when the above methods don't work:

```markdown
First line<br>
Second line
```

**Result:**
First line<br>
Second line

## Best Practices for This Project

### ✅ DO:
- Use double newlines for paragraph separation
- Use two trailing spaces + newline for line breaks within paragraphs
- Keep lines under 80-100 characters when possible
- Use consistent spacing around headers
- Use fenced code blocks with language specification

### ❌ DON'T:
- Use `<br>` tags unless absolutely necessary
- Mix different line break methods inconsistently
- Leave trailing spaces except for intentional line breaks
- Use excessive blank lines (more than 2 consecutive)

## Examples

### Headers
```markdown
# Main Title

## Section Header

### Subsection Header
```

### Lists
```markdown
- Item 1
- Item 2
  - Sub-item 2.1
  - Sub-item 2.2
- Item 3
```

### Code Blocks
```markdown
```bash
kubectl apply -f deployment.yaml
```
```

### Links and References
```markdown
See the [API documentation](./api-docs.md) for more details.

Or visit our [project website](https://example.com).
```

## Formatting Standards

1. **Headers**: Use ATX-style headers (`#`) with space after hash
2. **Emphasis**: Use `**bold**` and `*italic*` (not `__` or `_`)
3. **Code**: Use backticks for `inline code`
4. **Lists**: Use `-` for unordered lists, numbers for ordered
5. **Links**: Use reference-style links for repeated URLs

## Validation

To ensure consistency, consider using:
- markdownlint (CLI tool)
- Prettier (formatter)
- EditorConfig for consistent spacing

---

*This style guide ensures consistent, readable markdown across the go-queue project.*