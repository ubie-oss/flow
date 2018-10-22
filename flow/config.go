package flow

type Config struct {
	ManifestOwner string `yaml:"manifest_owner"`
	ManifestName  string `yaml:"manifest_name"`

	ApplicationList []Application `yaml:"applications"`
	GitAuthor       GitAuthor     `yaml:"git_author"`

	SlackNotifiyChannel string `yaml:"slack_notify_channel"`
}

type Application struct {
	Name        string   `yaml:"name"`
	SourceOwner string   `yaml:"source_owner"`
	SourceName  string   `yaml:"source_name"`
	BaseBranch  string   `yaml:"base_branch_name"`
	Env         string   `yaml:"env"`
	ImageName   string   `yaml:"image_tag"`
	Manifests   []string `yaml:"manifests"`
}

type GitAuthor struct {
	Name  string `yaml:"name"`
	Email string `yaml:"email"`
}
