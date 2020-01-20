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
	assert.Equal(t, "release/production-bar", r.Message)
	assert.Equal(t, "release/production-bar", r.CommitBranch)
	assert.Equal(t, "master", r.BaseBranch)

	manifest.BaseBranch = "production"
	r2 := newRelease(app, manifest, version)
	assert.Equal(t, "release/production-bar", r2.Message)
	assert.Equal(t, "release/production-bar", r2.CommitBranch)
	assert.Equal(t, "production", r2.BaseBranch)

	manifest.CommitWithoutPR = true
	r3 := newRelease(app, manifest, version)
	assert.Equal(t, "release/production-bar", r3.Message)
	assert.Equal(t, "production", r3.CommitBranch)
	assert.Equal(t, "production", r3.BaseBranch)

	app2 := Application{
		Name: "foo",
	}
	manifest2 := Manifest{
		Env:        "dev",
		BaseBranch: "dev",
	}

	r4 := newRelease(app2, manifest2, version)
	assert.Equal(t, "release/dev-bar", r4.Message)
	assert.Equal(t, "release/dev-bar", r4.CommitBranch)
	assert.Equal(t, "dev", r4.BaseBranch)

	manifest2.CommitWithoutPR = true
	r5 := newRelease(app2, manifest2, version)
	assert.Equal(t, "release/dev-bar", r5.Message)
	assert.Equal(t, "dev", r5.CommitBranch)
	assert.Equal(t, "dev", r5.BaseBranch)

}

func TestShouldProcess(t *testing.T) {
	// ignore empty and latest
	assert.Equal(t, false, shouldProcess(Manifest{}, ""))
	assert.Equal(t, false, shouldProcess(Manifest{}, "latest"))

	// usual tag
	assert.Equal(t, true, shouldProcess(Manifest{}, "foo"))

	// test include prefix
	m1 := Manifest{
		Filters: Filters{
			IncludePrefixes: []string{
				"v",
			},
		},
	}
	assert.Equal(t, true, shouldProcess(m1, "v123"))
	assert.Equal(t, false, shouldProcess(m1, "release-foo"))

	// test exclude prefix
	m2 := Manifest{
		Filters: Filters{
			ExcludePrefixes: []string{
				"v",
			},
		},
	}
	assert.Equal(t, false, shouldProcess(m2, "v123"))
	assert.Equal(t, true, shouldProcess(m2, "release-foo"))

	// mixed
	m3 := Manifest{
		Filters: Filters{
			IncludePrefixes: []string{
				"v",
			},
			ExcludePrefixes: []string{
				"vv",
			},
		},
	}
	assert.Equal(t, true, shouldProcess(m3, "v123"))
	assert.Equal(t, false, shouldProcess(m3, "vv123"))
	assert.Equal(t, false, shouldProcess(m3, "release-foo"))
}