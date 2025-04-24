# Commit Message Format Guide

This project follows the [Conventional Commits](mdc:https:/www.conventionalcommits.org/en/v1.0.0) specification for Git commit messages. This leads to more readable messages that are easy to follow when looking through the project history.

## Format

Each commit message consists of a **header**, a **body**, and a **footer**.

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

### Header

The header is mandatory and includes a **type**, an optional **scope**, and a **description**.

*   **Type:** Must be one of the following:
    *   `feat`: A new feature for the user.
    *   `fix`: A bug fix for the user.
    *   `build`: Changes that affect the build system or external dependencies (e.g., npm, Go modules).
    *   `chore`: Other changes that don't modify src or test files (e.g., updating dependencies, build tasks).
    *   `ci`: Changes to our CI configuration files and scripts.
    *   `docs`: Documentation only changes.
    *   `perf`: A code change that improves performance.
    *   `refactor`: A code change that neither fixes a bug nor adds a feature.
    *   `revert`: Reverts a previous commit.
    *   `style`: Changes that do not affect the meaning of the code (white-space, formatting, missing semi-colons, etc).
    *   `test`: Adding missing tests or correcting existing tests.

*   **Scope (Optional):** A noun describing the section of the codebase affected (e.g., `api`, `frontend`, `backend`, `auth`, `ui`, `parser`). Use lowercase.

*   **Description:** A short summary of the code changes.
    *   Use the imperative, present tense: "change" not "changed" nor "changes".
    *   Don't capitalize the first letter.
    *   No dot (`.`) at the end.

### Body (Optional)

*   Use the imperative, present tense.
*   Includes motivation for the change and contrasts with previous behavior.
*   Use blank lines to separate paragraphs.

### Footer (Optional)

*   Contains information about **Breaking Changes** and **references to issues** that this commit closes.
*   **Breaking Change:** Start with `BREAKING CHANGE:` followed by a description of the change, justification, and migration notes.
*   **Issue References:** Use keywords like `Closes`, `Fixes`, `Resolves` followed by the issue number (e.g., `Closes #123`, `Fixes #456, #789`).

## Examples

**Commit with scope, body, and footer:**

```
feat(api): add user registration endpoint

Implement the POST /users endpoint to allow new users to register.
Includes input validation and password hashing.

Closes #25
BREAKING CHANGE: The user ID format has changed from integer to UUID.
Update client-side code accordingly.
```

**Simple fix:**

```
fix: correct typo in documentation
```

**Refactor without scope:**

```
refactor: simplify internal data processing logic
```

**Commit affecting build:**

```
build: update Go version to 1.22
```

