package conventional

type (
	// Commit represents the components of a Conventional Commit message.
	// See https://www.conventionalcommits.org/ for the specification.
	Commit struct {
		// Type describes the kind of change (e.g., "feat", "fix", "docs").
		Type string
		// Type describes the kind of change (e.g., "feat", "fix", "docs").
		Scope string
		// Title is a short description of the change.
		Title string
		// Body is an optional longer description.
		Body string
		// Footer is optional metadata (e.g., "BREAKING CHANGE:", issue refs).
		Footer string
	}
)
