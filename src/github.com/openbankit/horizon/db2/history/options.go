package history

const (
	OPTIONS_MAX_REVERSAL_DURATION string = "max_reversal_duration"
)

type Options struct {
	Name string `db:"name"`
	Data string `db:"data"`
}

func (o *Options) MaxReversalDuration() *MaxReversalDuration {
	result := MaxReversalDuration(*o)
	return &result
}
