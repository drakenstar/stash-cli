---
description: Create a git commit
---

Create a git commit for the current worktree changes.

Follow the repository git safety rules before committing: inspect `git status`,
the full staged and unstaged diff, and recent commit messages before deciding
what to stage and how to write the message. Do not commit secrets or unrelated
changes.

Commit message rules:

- Subject must be no longer than 50 characters.
- Subject must not start with an uppercase letter.
- Subject should use a short command style that indicates the focus of the
  change, such as `update command parsing` or `fix gallery filtering`.
- Include an extended commit message with one short rationale paragraph.
- Wrap the extended message at 80 characters.

After committing, run `git status` and report the commit hash and remaining
worktree state.
