package conventional

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

type (
	testCases struct {
		commitData *Commit
		err        error
		result     string
	}
)

var buildCommitMessageTestCases = []testCases{
	{
		result: "feat(api): some changes\n\nChanges description\n",
		commitData: &Commit{
			Title: "some changes",
			Type:  "feat",
			Scope: "api",
			Body:  "Changes description",
		},
		err: nil,
	},
	{
		result: "feat(api): some changes\n",
		commitData: &Commit{
			Title: "some changes",
			Type:  "feat",
			Scope: "api",
		},
		err: nil,
	},
	{
		result: "feat: some changes\n\nChanges description\n",
		commitData: &Commit{
			Title: "some changes",
			Type:  "feat",
			Body:  "Changes description",
		},
		err: nil,
	},
	{
		result: "feat(api): some changes\n\nChanges description\n\nIssue: TEST-1\n",
		commitData: &Commit{
			Title: "some changes",
			Type:  "feat",
			Scope: "api",
			Body:  "Changes description",
			Issue: "TEST-1",
		},
		err: nil,
	},
	{
		result: "feat(api)!: some changes\n\nChanges description\n\nBREAKING CHANGE: Something breaking\nIssue: TEST-1\n",
		commitData: &Commit{
			Title:          "some changes",
			Type:           "feat",
			Scope:          "api",
			Body:           "Changes description",
			Issue:          "TEST-1",
			BreakingChange: "Something breaking",
		},
		err: nil,
	},
	{
		result: "feat!: some changes\n\nChanges description\n\nBREAKING CHANGE: Something breaking\n",
		commitData: &Commit{
			Title:          "some changes",
			Type:           "feat",
			Body:           "Changes description",
			BreakingChange: "Something breaking",
		},
		err: nil,
	},
	{
		result: "",
		commitData: &Commit{
			Type: "feat",
			Body: "Changes description",
		},
		err: ErrRequiredPartNotPreset,
	},
	{
		result: "",
		commitData: &Commit{
			Title: "some changes",
			Body:  "Changes description",
		},
		err: ErrRequiredPartNotPreset,
	},
}

func TestBuildCommitMessage(t *testing.T) {
	t.Parallel()
	for i, tt := range buildCommitMessageTestCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()

			message, err := BuildCommitMessage(tt.commitData)

			assert.Equal(t, tt.result, message)
			assert.ErrorIs(t, err, tt.err)
		})
	}
}
