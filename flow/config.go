package flow

type Config struct {
	ManifestOwner      string `yaml:"manifest_owner"`
	ManifestName       string `yaml:"manifest_name"`
	ManifestBaseBranch string `yaml:"manifest_base_branch"`

	ApplicationList []Application `yaml:"applications"`
	GitAuthor       GitAuthor     `yaml:"git_author"`

	SlackNotifiyChannel string `yaml:"slack_notify_channel"`
}

type Application struct {
	Name        string     `yaml:"name"`
	SourceOwner string     `yaml:"source_owner"`
	SourceName  string     `yaml:"source_name"`
	Env         string     `yaml:"env"`
	ImageName   string     `yaml:"image_tag"`
	Manifests   []Manifest `yaml:"manifests"`
}

type Manifest struct {
	Env   string   `yaml:"env"`
	Files []string `yaml:"files"`
}

type GitAuthor struct {
	Name  string `yaml:"name"`
	Email string `yaml:"email"`
}
