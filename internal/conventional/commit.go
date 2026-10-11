package conventional

type (
	// Commit represents the components of a Conventional Commit message.
	// See https://www.conventionalcommits.org/ for the specification.
	Commit struct {
		// Type describes the kind of change (e.g., "feat", "fix", "docs").
		Type string `json:"type"`
		// Scope is an optional context for the change (e.g., "api", "cli").
		Scope string `json:"scope,omitempty"`
		// Title is a short description of the change.
		Title string `json:"title"`
		// Body is an optional longer description.
		Body string `json:"body,omitempty"`
		// BreakingChange describes breaking changes introduced by the commit.
		BreakingChange string `json:"breakingChange,omitempty"`
		// Issue references the issue closed by this commit.
		Issue string `json:"issue,omitempty"`
	}
)
