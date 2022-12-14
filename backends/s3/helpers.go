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

type pair struct {
	key   string
	value string
}

func collectMap(writtenChan chan pair) map[string]string {

	written := map[string]string{}
	for {
		select {
		case pair := <-writtenChan:
			written[pair.key] = pair.value
		default:
			close(writtenChan)
			return written
		}
	}
}

func collectArray(writtenChan chan string) []string {

	written := []string{}
	for {
		select {
		case val := <-writtenChan:
			written = append(written, val)
		default:
			close(writtenChan)
			return written
		}
	}
}
