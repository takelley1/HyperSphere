# Repository Guidelines

YOU MUST FOLLOW ALL THESE RULES WHENEVER MAKING ANY CODE CHANGES!

## Goals
- Your goal is to reach parity with the k9s project (in the current repo), except adapted to VMWare's vSphere / vCenter, instead of Kubernetes
- as you work, add more items to DESIGN.md, classify in terms of long-term, mediu-term, and short-term goals. then add sub-tasks under each. as you compelte them, remove them

## General
- This project is critical -- please focus!
- Don't be obsequious, sycophantic, excessively apologetic, overly verbose, or overly polite.
- Be concise—omit pleasantries, avoid repeating the prompt, and skip redundant scaffolding.
- There should always be a single canonical implementation of everything in the code.
- Implement TDD. When I give you a task, write a test, ensure the test fails, then write code to make the test pass. Once the code makes the test pass, keep the test and the code.
- Ensure test coverage STAYS at 100%. If you add anything and coverage drops below 100%, add more tests to bring coverage back to 100%.

## Planning
- Never alter the core tech stack without my explicit approval. Only add dependencies if I give you permission in a prompt.
- Think step-by-step before making a change.
- For large changes, provide an implementation plan.
- Before making any changes or implementing anything, check if a pattern for it exists already in the codebase, and follow that pattern.

## Code Style
- Target Go.
- Prefer configuration using environment variables and CLI options unless stated otherwise.
- Always prioritize the simplest solution over complexity.
- Code must be easy to read and understand.
- Keep code as simple as possible. Avoid unnecessary complexity.
- Follow DRY and YAGNI coding principles.
- Follow SOLID principles (e.g., single responsibility, dependency inversion) where applicable.
- DO NOT over-engineer code!
- Never duplicate code.

## Formatting
- Ruff formatting is used. Never try to undo formatting done by Ruff.
- Never modify config/ruff.toml
- Keep lines under 100 characters in length.
- Ensure all lines DO NOT have trailing whitespace.

## Variables
- Use meaningful names for variables, functions, etc. Names should reveal intent. Don't use short names for variables.

## Imports
- Ensure all imports are at the top of the file. Don't import modules inside functions.
- Ensure third party imports always come after builtin imports.
- Ensure all imports are actually used.
- Only do a non-toplevel import if it's absolutely necessary.

## Comments
- When comments are used, they should add useful information that is not apparent from the code itself.
- Comments should be full, gramatically-correct sentences with punctuation.
- Don't use inline comments. Instead, put the comment on the line before the relevant code.
- At the top of every source code file, ensure there's a comment with its full path relative to the repo root and a description of its function, like this:
  ```
  # Path: ssherlock/tests/test_ssherlock_server_views.py
  # Description: Validate SSHerlock server view flows, responses, and integrations.
  ```

## Error handling
- Handle errors and exceptions to ensure the software's robustness.
- Don't catch overly-broad exceptions. Instead, catch specific exceptions.
- Use exceptions rather than error codes for handling errors.
    - However, don't be overly-defensive.
- Don't create alternative or backups paths for doing something.
- There should always be one main canonical path for doing something. Don't be overly-defensive and code additional backup code paths that implement core functionality.

## Functions
- Functions should be small and do one thing.
- YOU SHOULD ABSOLUTELY keep functions under 30 lines.
- Before making a change to a function, check if it would increase the size of a function to >30 lines. If so, you should refactor first.
- Proactively refactor.
- Function names should describe what they do.
- Prefer fewer arguments in functions. Aim for less than about 7.

## Tests
- Follow test-driven-development (TDD) principles. When a feature is described, write a failing test for it first, then make the test pass with the minimum amount of code necessary.
- If you need to write more code than what's necessary to make the failing test pass, then that means you need to write more test code. This adheres to TDD.
- Obviously if you're fixing tests or writing tests themselves, you don't need to follow TDD for that. TDD is only for adding features to the application code or fixing bugs to prevent regressions.
- Include comprehensive tests for major features; suggest edge case tests (e.g. invalid inputs).
- Include unit and integration tests.
- Don't fix failing tests without my express and explicit permission. If a change breaks tests, STOP and ask me for guidance on what to do.

## Commits
- Follow the existing `area: short summary` convention (for example, `tests: add runner fixtures`); limit the subject to about 100 characters.
- Be descriptive in the body of commit messages, describing exactly what was implemented. Don't include newlines in commit bodies.
- Before committing, run the linting and tests in the ./scripts dir to verify functionality. Fix everything mentioned in the lint script. The lint scripts will exit 0 every time, so make sure you inspect the actual output.
- So after you complete my instructions and implement a feature, fix a bug, cleanup the code, or improve tests, run the lint scripts, the test scripts, fix issues, re-run the scripts until they're clean, and then you can commit your changes with git.

## Documentation
- After each component, summarize what’s done in a CHANGELOG.md file.
- Use the `date` command to obtain the correct date first before writing to the CHANGELOG.md file.
- Don't touch the USER_GUIDE.md or DEPLOYMENT_GUIDE.md files unless told to do so.
- Use the .scratchpad.txt file for temporary storage, plans, and managing your own memory.
- Ignore the TODO.md file unless explicitly told to reference it, that's for humans to use.

## Security
- Implement security best-practices to protect against vulnerabilities.
- Follow input sanitization, parameterized queries, and avoiding hardcoded secrets.
- Follow web server design best practices for security.

## For bash/zsh/fish code only
- Follow all shellcheck conventions and rules.
- Handle errors gracefully.
- Use `/usr/bin/env bash` in the shebang line.
- Use `set -euo pipefail`.
- Use `[[ ]]` instead of `[ ]`.
- Use `"$()"` instead of `` ``.
- Use `"${VAR}"` instead of `"$VAR"`.
- Don't use arrays unless absolutely necessary.
- Use `printf` instead of `echo`.
- Encapsulate functionality in functions.

## Examples

<Shell>
    - Correct shebang example:
        <example>
        #!/usr/bin/env bash
        </example>

    - Correct shell options example:
        <example>
        set -euo pipefail
        </example>

    - Correct if-statement formatting example:
        <example>
        if [[ -z "${URL}" ]]; then
          exit 1
        fi
        </example>

    - Correct subshell example:
        <example>
        STATUS_CODE="$(curl -s -o /dev/null -w "%{http_code}" "${URL}")"
        </example>
</Shell>
