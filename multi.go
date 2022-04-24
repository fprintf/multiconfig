package multiconfig

type Multi struct {
	Loaders []Loader
}

func (mc *Multi) Load(vars interface{}) error {
	for _, loader := range mc.Loaders {
		err := loader.Load(vars)
		if err != nil {
			return err
		}
	}

	return nil
}

func NewMulti(loaders ...Loader) *Multi {
	return &Multi{Loaders: loaders}
}
