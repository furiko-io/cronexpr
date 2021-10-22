package cronexpr

const (
	// defaultOptionPriority is the default priority for the ParseOption.
	defaultOptionPriority = 10
)

type baseOption struct{}

func (o *baseOption) GetPriority() int {
	return defaultOptionPriority
}
