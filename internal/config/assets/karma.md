Generate a Git commit message following the Karma (Angular) commit message conventions.

Rules:
- Use the format:
  <type>(<scope>): <subject>

  <blank line>
  <body>

  <blank line>
  <footer>

- Allowed types:
  feat, fix, docs, style, refactor, perf, test, build, ci, chore, revert

- Scope is required and should be a noun describing the affected area
  (e.g. parser, compiler, http, auth, cli)

- Use the imperative mood in the subject (e.g. "add", "fix", "change")
- Do not capitalize the first letter of the subject
- Do not end the subject with a period
- Keep the subject concise (preferably under 50 characters)

- The body should explain the motivation and reasoning for the change,
  not just what was changed
- Wrap body lines at approximately 72 characters
- Separate paragraphs in the body with blank lines

- The footer is required when applicable and must use this format:
  BREAKING CHANGE: <description>
  or
  Closes #<issue number>

- If the change introduces a breaking change:
  - Use BREAKING CHANGE in the footer
  - The subject must describe the change, not the migration steps

- Do not reference the diff, file names, or line numbers directly
- Do not wrap the message in backticks or code fences
- Output only the commit message, without any explanations or metadata

Write the commit message based on the following git diff:
