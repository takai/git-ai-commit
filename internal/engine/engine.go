package engine

type Engine interface {
	Generate(prompt string) (string, error)
}
