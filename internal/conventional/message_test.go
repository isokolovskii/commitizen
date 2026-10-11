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

const (
	testTitle    = "some changes"
	testType     = "feat"
	testScope    = "api"
	testBody     = "Changes description"
	testIssue    = "TEST-1"
	testBreaking = "Something breaking"
)

var buildCommitMessageTestCases = []testCases{
	{
		result: "feat(api): some changes\n\nChanges description\n",
		commitData: &Commit{
			Title: testTitle,
			Type:  testType,
			Scope: testScope,
			Body:  testBody,
		},
		err: nil,
	},
	{
		result: "feat(api): some changes\n",
		commitData: &Commit{
			Title: testTitle,
			Type:  testType,
			Scope: testScope,
		},
		err: nil,
	},
	{
		result: "feat: some changes\n\nChanges description\n",
		commitData: &Commit{
			Title: testTitle,
			Type:  testType,
			Body:  testBody,
		},
		err: nil,
	},
	{
		result: "feat(api): some changes\n\nChanges description\n\nIssue: TEST-1\n",
		commitData: &Commit{
			Title: testTitle,
			Type:  testType,
			Scope: testScope,
			Body:  testBody,
			Issue: testIssue,
		},
		err: nil,
	},
	{
		result: "feat(api)!: some changes\n\nChanges description\n\nBREAKING CHANGE: Something breaking\nIssue: TEST-1\n",
		commitData: &Commit{
			Title:          testTitle,
			Type:           testType,
			Scope:          testScope,
			Body:           testBody,
			Issue:          testIssue,
			BreakingChange: testBreaking,
		},
		err: nil,
	},
	{
		result: "feat!: some changes\n\nChanges description\n\nBREAKING CHANGE: Something breaking\n",
		commitData: &Commit{
			Title:          testTitle,
			Type:           testType,
			Body:           testBody,
			BreakingChange: testBreaking,
		},
		err: nil,
	},
	{
		result: "",
		commitData: &Commit{
			Type: testType,
			Body: testBody,
		},
		err: ErrRequiredPartNotPreset,
	},
	{
		result: "",
		commitData: &Commit{
			Title: testTitle,
			Body:  testBody,
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
