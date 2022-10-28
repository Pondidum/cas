package s3

import "github.com/hashicorp/go-multierror"

func collectErrors(errChan chan error) error {

	errors := []error{}
	for {
		select {
		case e := <-errChan:
			errors = append(errors, e)
		default:
			close(errChan)

			if len(errors) > 0 {
				return multierror.Append(nil, errors...)
			}

			return nil
		}
	}
}
