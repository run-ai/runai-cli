package util


// Parallel - running multipule functions in parallel 
func Parallel( fns ...func() error) error {
	errs := make(chan error)
	for _, fn := range fns {
		thisFn := fn
		go func() {
			errs <- thisFn()
		}()
	}

	for count := 0; count < len(fns); count++ {
		select {
		case e := <-errs:
			if e != nil {
				return e
			}
		}
	}

	return nil
}