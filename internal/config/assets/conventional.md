Generate a Git commit message following the Conventional Commits specification.

Rules:
- Use the format: <type>(<optional scope>): <subject>
- Allowed types: feat, fix, docs, style, refactor, perf, test, build, ci, chore
- Use the imperative mood in the subject (e.g. "add", "fix", "update")
- Keep the subject concise (ideally under 50 characters)
- Do not end the subject with a period
- Include a scope only if it adds clarity (e.g. module, package, feature name)
- If the change introduces a breaking change:
  - Add "!" after the type or scope
  - Describe the breaking change clearly in the footer using "BREAKING CHANGE:"
- If useful, add a body explaining *why* the change was made, not just what changed
- Wrap body lines at approximately 72 characters
