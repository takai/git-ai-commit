Generate a Git commit message following the official gitmoji specification
(https://gitmoji.dev/).

Rules:
- Start the commit message with exactly one gitmoji emoji that represents the
  purpose of the change, as defined in the official gitmoji list
  (e.g. ✨ introduce new features, 🐛 fix bugs, 📝 add or update documentation,
  ♻️ refactor code, ⚡ improve performance, 🔧 change configuration)

- The first line must follow this format:
  <emoji> <subject>

- Subject rules:
  - Use the imperative mood (e.g. "add", "fix", "update", "remove")
  - Describe what the change does, not what you did
  - Keep it concise (preferably under 50 characters)
  - Do not end with a period
  - Use lowercase, unless a proper noun is required

- Body (optional):
  - Add a blank line after the subject before the body
  - Use the body to explain *why* the change was made or provide important context
  - Avoid repeating the subject line
  - Wrap lines at approximately 72 characters

- Use only one emoji per commit
- Do not include Conventional Commits types (feat, fix, etc.)
- Do not mix multiple conventions in a single commit
- Do not mention the diff, file names, or line numbers explicitly
- Output only the commit message, with no explanations or metadata

Write the commit message based on the following git diff:
