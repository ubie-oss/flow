package flow

import (
	"fmt"
	"os"
)

func (f *Flow) process(e Event) error {
	if e.RepoName != cfg.ManifestName {
		f.createPR()
		f.notifyBuild()
	} else {
		f.notifyDeploy()
	}

	return nil
}

func (f *Flow) createPR() {
	fmt.Fprintf(os.Stdout, "@todo create pr")
}

func (f *Flow) notifyBuild() {
	fmt.Fprintf(os.Stdout, "@todo notify build result")
}

func (f *Flow) notifyDeploy() {
	fmt.Fprintf(os.Stdout, "@todo notify deploy result")
}
