# Markdown Formatting Guide

This guide addresses the question from PR #21: "in md file do i always add <br> tag to change line"

## Answer: No, you typically don't need `<br>` tags in markdown

## Proper Markdown Line Break Formatting

### 1. Paragraph Breaks (Recommended)
Use an empty line between paragraphs:
```markdown
This is paragraph one.

This is paragraph two.
```

### 2. Soft Line Breaks (Use Sparingly)
For a line break within a paragraph, use two spaces at the end of the line:
```markdown
First line  
Second line (same paragraph)
```

### 3. Avoid HTML `<br>` Tags
While `<br>` tags work in markdown, they are not the standard approach:
```markdown
<!-- Avoid this -->
First line<br>
Second line

<!-- Prefer this -->
First line

Second line
```

## Why This Matters

1. **Consistency**: Standard markdown formatting ensures consistency across the repository
2. **Readability**: Clean markdown is easier to read in both source and rendered form
3. **Compatibility**: Standard formatting works better with markdown processors and linters

## What We Fixed

- Removed trailing spaces that were unnecessarily creating line breaks
- Ensured consistent formatting across all markdown files
- No `<br>` tags were needed or used in this repository

## Best Practices for This Repository

1. Use empty lines for paragraph separation
2. Avoid trailing spaces unless intentionally creating soft line breaks
3. Keep markdown simple and readable
4. Let the content structure guide the formatting