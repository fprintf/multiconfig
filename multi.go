package multiconfig

type MultiLoader struct {
	Loaders []Loader
}

func (mc *MultiLoader) Load(vars interface{}) error {
	for _, loader := range mc.Loaders {
		err := loader.Load(vars)
		if err != nil {
			return err
		}
	}

	return nil
}

func NewMultiLoader(loaders ...Loader) *MultiLoader {
	return &MultiLoader{Loaders: loaders}
}
