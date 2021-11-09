package flow

type Config struct {
	ApplicationList []Application `yaml:"applications"`
	GitAuthor       GitAuthor     `yaml:"git_author"`

	DefaultManifestOwner string `yaml:"default_manifest_owner"`
	DefaultManifestName  string `yaml:"default_manifest_name"`

	SlackNotifiyChannel string `yaml:"slack_notify_channel"`
	DefaultBranch       string `yaml:"default_branch"`
}

type Application struct {
	SourceOwner        string `yaml:"source_owner"`
	SourceName         string `yaml:"source_name"`
	ManifestOwner      string `yaml:"manifest_owner"`
	ManifestName       string `yaml:"manifest_name"`
	ManifestBaseBranch string `yaml:"manifest_base_branch"`

	RewriteNewTag           bool     `yaml:"rewrite_new_tag"`
	AdditionalRewriteKeys   []string `yaml:"additional_rewrite_keys"`
	AdditionalRewritePrefix []string `yaml:"additional_rewrite_prefix"`

	Image     string     `yaml:"image"`
	Manifests []Manifest `yaml:"manifests"`
}

type Manifest struct {
	Env                           string   `yaml:"env"`
	Name                          string   `yaml:"name"`
	ShowSourceOwner               bool     `yaml:"show_source_owner"`
	HideSourceName                bool     `yaml:"hide_source_name"`
	HideSourceReleaseDesc         bool     `yaml:"hide_source_release_desc"`
	HideSourceReleasePullRequests bool     `yaml:"hide_source_release_pull_requests"`
	Files                         []string `yaml:"files"`
	Filters                       Filters  `yaml:"filters"`
	PRBody                        string   `yaml:"pr_body"`
	BaseBranch                    string   `yaml:"base_branch"`
	CommitWithoutPR               bool     `yaml:"commit_without_pr"`
	Labels                        []string `yaml:"labels"`
}

type Filters struct {
	IncludePrefixes []string `yaml:"include_prefixes"`
	ExcludePrefixes []string `yaml:"exclude_prefixes"`
}

type GitAuthor struct {
	Name  string `yaml:"name"`
	Email string `yaml:"email"`
}
