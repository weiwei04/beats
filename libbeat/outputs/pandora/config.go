package pandora

type config struct {
	Endpoint   string `config:"endpoint"`
	AK         string `config:"ak"`
	SK         string `config:"sk"`
	Region     string `config:"region"`
	Batch      int    `config:"batch" validate:"min=1"`
	MaxRetries int    `config:"max_retries"`
}

var (
	defaultConfig = config{
		Batch:      10,
		MaxRetries: 3,
	}
)
