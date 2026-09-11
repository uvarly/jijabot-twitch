package phrasebook

import (
	"bytes"
	"errors"
	"fmt"
	"math/rand/v2"
)

var (
	ErrCommandNotFound  = errors.New("command not found")
	ErrScenarioNotFound = errors.New("scenario not found")
)

type RNG interface {
	Float64() float64
}

type MathRandRNG struct{}

func (MathRandRNG) Float64() float64 { return rand.Float64() }

type Option func(*Picker)

func WithRNG(rng RNG) Option {
	return func(p *Picker) { p.rng = rng }
}

type Picker struct {
	book *Phrasebook
	rng  RNG
}

func NewPicker(book *Phrasebook, options ...Option) *Picker {
	picker := &Picker{
		book: book,
		rng:  MathRandRNG{},
	}

	for _, o := range options {
		o(picker)
	}

	return picker
}

func (p *Picker) Pick(command, scenario string, data any) (string, error) {
	scenarios, ok := p.book.entries[command]
	if !ok {
		return "", fmt.Errorf("failed to find command %q: %w", command, ErrCommandNotFound)
	}

	phrases, ok := scenarios[scenario]
	if !ok {
		return "", fmt.Errorf("failed to find scenario %q: %w", scenario, ErrScenarioNotFound)
	}

	chosen := pickWeighted(phrases, p.rng)

	var buf bytes.Buffer

	if err := chosen.Text.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template %s.%s: %w", command, scenario, err)
	}

	return buf.String(), nil
}

func pickWeighted(phrases []phrase, rng RNG) phrase {
	var (
		totalWeight      float64
		cumulativeWeight float64
	)

	for _, phrase := range phrases {
		totalWeight += phrase.Weight
	}

	randomFloat64 := rng.Float64() * totalWeight

	for _, phrase := range phrases {
		cumulativeWeight += phrase.Weight

		if randomFloat64 < cumulativeWeight {
			return phrase
		}
	}

	return phrases[len(phrases)-1]
}
