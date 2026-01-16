Generate a Git commit message following Git's official commit message guidelines.

Rules:
- Use a short summary line followed by an optional detailed body

- Summary line:
  - Use the imperative mood (e.g. "add", "fix", "update")
  - Describe what the commit does, not what you did
  - Keep the summary concise (ideally under 50 characters)
  - Capitalize the first letter
  - Do not end the summary with a period

- After the summary, insert a blank line

- Body (optional but recommended for non-trivial changes):
  - Explain the motivation and reasoning behind the change
  - Describe why the change was necessary and what problem it solves
  - Avoid repeating the summary
  - Wrap lines at approximately 72 characters
  - Use full sentences and paragraphs as needed

- Do not include metadata such as timestamps or author names
- Do not mention the diff, file names, or line numbers explicitly
- Output only the commit message, with no additional explanations

Write the commit message based on the following git diff:
