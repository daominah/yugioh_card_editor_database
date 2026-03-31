# Code Style

- Use explicit `if`/`else` blocks; avoid ternary expressions.

# Writing Style

- Avoid dashes (`-` or `—`) in the middle of sentences;
  rephrase or use colons instead.
  Compound words are allowed: "real-time", "back-end".
- Use standard straight quotes (`'` and `"`) instead of curly quotes.

# Testing

Do not weaken or remove test assertions when the error looks like an environment issue;
ask the user to fix the environment setup instead.

# Skills

- `commit-messages`: ALWAYS invoke this skill before ANY git commit,
  including short requests like "commit", "git commit", "commit push", "git commit push", etc.
