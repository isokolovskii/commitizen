package conventional

type (
	// Commit represents the components of a Conventional Commit message.
	// See https://www.conventionalcommits.org/ for the specification.
	Commit struct {
		// Type describes the kind of change (e.g., "feat", "fix", "docs").
		Type string
		// Scope is an optional context for the change (e.g., "api", "cli").
		Scope string
		// Title is a short description of the change.
		Title string
		// Body is an optional longer description.
		Body string
		// Breaking change description.
		BreakingChange string
		// Issue number.
		Issue string
	}
)
