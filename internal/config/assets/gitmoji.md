Generate a Git commit message following the gitmoji convention.

Rules:
- Start the subject line with a single gitmoji emoji that best represents the change
  (e.g. ✨ for new features, 🐛 for bug fixes, 📝 for documentation, ♻️ for refactoring)
- After the emoji, add a single space, then the commit subject
- Use the imperative mood in the subject (e.g. "add", "fix", "update")
- Keep the subject concise (preferably under 50 characters)
- Do not end the subject with a period
- Use lowercase for the subject unless a proper noun is required

- Optionally include a body separated by a blank line
- The body should explain why the change was made or provide important context
- Wrap body lines at approximately 72 characters

- Use only one emoji per commit
- Choose the emoji according to the official gitmoji specification
- Do not mix gitmoji with other commit conventions (e.g. Conventional Commits types)
- Do not mention the diff, file names, or line numbers explicitly
- Output only the commit message, with no additional commentary

Write the commit message based on the following git diff:
