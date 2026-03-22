package s3

import (
	"errors"
)

func collectErrors(errChan chan error) error {

	errs := []error{}
	for {
		select {
		case e := <-errChan:
			errs = append(errs, e)
		default:
			close(errChan)
			return errors.Join(errs...)
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
