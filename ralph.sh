#!/usr/bin/env bash

../ralph-wiggum-codex/bin/ralph loop \
\
"Read AGENTS.md
Implement the next unfulfilled requirement in REQUIREMENTS.md file.
Write a failing test for the requirement, then get the test to pass.
Once the requirement has been satisfied, commit and then mark the requirement as fulfilled.
Refactor if needed.
If the project structure is disorganized or need organization, do that first.
If the worktree is dirty, commit first before making any changes.
Make sure you use git to commit your changes." \
--max-iterations 20 \
--completion-promise "<promse>DONE</promse>" \
-- --model gpt-5.3-codex -s workspace-write
