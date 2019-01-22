package flow

type Config struct {
	ApplicationList []Application `yaml:"applications"`
	GitAuthor       GitAuthor     `yaml:"git_author"`

	SlackNotifiyChannel string `yaml:"slack_notify_channel"`
}

type Application struct {
	Name string `yaml:"name"`

	TriggerID string `yaml:"trigger_id"`

	SourceOwner        string `yaml:"source_owner"`
	SourceName         string `yaml:"source_name"`
	ManifestOwner      string `yaml:"manifest_owner"`
	ManifestName       string `yaml:"manifest_name"`
	ManifestBaseBranch string `yaml:"manifest_base_branch"`

	ImageName string     `yaml:"image_tag"`
	Manifests []Manifest `yaml:"manifests"`
}

type Manifest struct {
	Env        string   `yaml:"env"`
	Files      []string `yaml:"files"`
	Filters    Filters  `yaml:"filters"`
	PRBody     string   `yaml:"pr_body"`
	BaseBranch string   `yaml:"base_branch"`
}

type Filters struct {
	IncludePrefixes []string `yaml:"include_prefixes"`
	ExcludePrefixes []string `yaml:"exclude_prefixes"`
}

type GitAuthor struct {
	Name  string `yaml:"name"`
	Email string `yaml:"email"`
}
