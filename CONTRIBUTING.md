# Contributing to Trident

First off, thank you for considering contributing to Trident! It's people like you that make open source a vibrant community. We welcome any form of contribution, not just code.

This document provides guidelines for contributing. Please feel free to propose changes to this document in a pull request.

## Code of Conduct

This project and everyone participating in it is governed by the [Trident Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code. Please report unacceptable behavior to at least one of the authors.

## How Can I Contribute?

### Reporting Bugs

Bugs are tracked as GitHub issues. Before opening a new issue, please perform a quick search to see if the problem has already been reported.

### Suggesting Enhancements

We love new ideas! If you have a suggestion for an enhancement, please open an issue with the `enhancement` label. Clearly describe the enhancement: what it is, why it's needed, and what the use case is.

### Your First Code Contribution

Unsure where to begin? You can start by looking through these `good first issue` and `help wanted` issues:

*   **Good first issues** - Issues that are well-defined and a good way to get familiar with the codebase.
*   **Help wanted** - Issues that are more involved and require a bit more context.

## Setting Up Your Development Environment

To contribute to Trident, you'll need a working Go development environment. We recommend using the latest stable version of Go.

You will also need:
*   `golangci-lint` for linting your code.
*   `pre-commit` for managing and maintaining our pre-commit hooks.

### Installing and Using Pre-commit Hooks

**This is a required step.** We use pre-commit hooks to automatically format code, check for issues, and validate commit messages before you even commit. This saves everyone time by catching errors early.

1.  **Install the `pre-commit` framework.**
    ```bash
    # Using pip (recommended)
    pip install pre-commit

    # Using Homebrew on macOS
    brew install pre-commit
    ```

2.  **Install the hooks in your local repository.** Run this command from the root of the `trident` project:
    ```bash
    pre-commit install
    ```

Now, every time you run `git commit`, the hooks defined in our `.pre-commit-config.yaml` file will run automatically. If a hook fails, it will report the error. Simply fix the issue (the hook might even fix it for you!), `git add` the modified files, and try to commit again.

## Pull Request Process

We follow the standard "Fork & Pull" GitHub workflow.

1.  **Fork** the repo and create your branch from `main`. A good branch name is descriptive, like `fix/user-auth-bug` or `feature/new-export-format`.

2.  Make your changes in your forked repository. As you work, your pre-commit hooks will ensure your code stays clean.

3.  If you need to run checks manually, you can use these commands:
    *   **Test:** `go test -race -cover ./...`
    *   **Lint:** `golangci-lint run`

4.  Commit your changes using the guidelines below.

### Commit Message Guidelines

**We enforce the [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) specification.** This helps us automate changelogs and makes the project history easy to read. Our pre-commit hooks will validate your commit message format.

The basic format is:
```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

**Common types:**
*   `feat`: A new feature for the user.
*   `fix`: A bug fix for the user.
*   `docs`: Changes to documentation only.
*   `style`: Code style changes (formatting, etc.) that do not affect the meaning of the code.
*   `refactor`: A code change that neither fixes a bug nor adds a feature.
*   `test`: Adding missing tests or correcting existing tests.
*   `chore`: Changes to the build process or auxiliary tools and libraries.

**Example Commits:**
```
fix: correct handling of nil pointers in user service
```
```
feat(api): add endpoint for user profile deletion

This introduces a new DELETE /api/v1/users/{id} endpoint.
It is protected by admin-level authentication.

Closes #42
```
```
docs: update installation instructions in README
```

### Developer Certificate of Origin (DCO)

We require all commits to be "signed off" using the DCO. This certifies that you wrote the code or otherwise have the right to contribute it. The pre-commit hooks will also check for this.

You can sign off your commit automatically by using the `-s` or `--signoff` flag:
```bash
git commit -s
```
This will add a `Signed-off-by: Your Name <your.email@example.com>` line to your commit message. Please use your real name.

5.  Push your branch to your fork on GitHub.

6.  Open a **Pull Request** to the `main` branch of the `trident` repository. Describe your changes and link to any relevant issues.

Once your PR is open, a maintainer will review it. Thank you for your contribution!
