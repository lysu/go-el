package patcher

// Path to Patch
type Path string

// Patch contains a group path and value
type Patch map[Path]interface{}

// Patcher use to patch in memory struct with path
type Patcher struct{}

// PatchIt do patch work
func (p *Patcher) PatchIt(target interface{}, patch Patch) error {

	for path, value := range patch {

		targetValue, err := Locate(target, path)
		if err != nil {
			return err
		}

		err = targetValue.SetValue(value)
		if err != nil {
			return err
		}

	}

	return nil
}
