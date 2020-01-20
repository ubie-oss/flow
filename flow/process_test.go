package flow

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRelease(t *testing.T) {
	cfg = &Config{
		GitAuthor: GitAuthor{
			Name:  "test",
			Email: "test@test.test",
		},
	}

	app := Application{
		Name:               "foo",
		ManifestBaseBranch: "master",
	}
	manifest := Manifest{
		Env: "production",
	}
	version := "bar"

	r := newRelease(app, manifest, version)
	assert.Equal(t, "release/production-bar", r.CommitBranch)
	assert.Equal(t, "master", r.BaseBranch)

	manifest.BaseBranch = "production"
	r2 := newRelease(app, manifest, version)
	assert.Equal(t, "release/production-bar", r2.CommitBranch)
	assert.Equal(t, "production", r2.BaseBranch)

	manifest.CommitWithoutPR = true
	r3 := newRelease(app, manifest, version)
	assert.Equal(t, "production", r3.CommitBranch)
	assert.Equal(t, "production", r3.BaseBranch)
}

func TestShouldProcess(t *testing.T) {
	assert.Equal(t, false, shouldProcess(Manifest{}, ""))
	assert.Equal(t, false, shouldProcess(Manifest{}, "latest"))

	// @todo test filters
}
