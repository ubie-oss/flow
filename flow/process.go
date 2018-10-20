package flow

import (
	"fmt"
	"os"
)

func (f *Flow) process(e Event) error {
	if e.isSuuccess() { // Cloud Build Success

		if e.isApplicationBuild() { // Build for Application
			prURL, err := f.createRelasePR(e)
			if err != nil {
				f.notifyFalure(e)
				return err
			}
			return f.notifyRelasePR(e, prURL)
		}

		// Build for Deployment
		return f.notifyDeploy(e)
	}

	// Code Build Failure
	return f.notifyFalure(e)
}

func (f *Flow) createRelasePR(e Event) (string, error) {
	fmt.Fprintf(os.Stdout, "@todo create pr\n")

	return "", nil
}

func (f *Flow) notifyRelasePR(e Event, prURL string) error {
	fmt.Fprintf(os.Stdout, "@todo notify build result\n")
	return nil
}

func (f *Flow) notifyDeploy(e Event) error {
	fmt.Fprintf(os.Stdout, "@todo notify deploy result\n")
	return nil
}

func (f *Flow) notifyFalure(e Event) error {
	fmt.Fprintf(os.Stdout, "@todo notify failure\n")
	return nil
}
