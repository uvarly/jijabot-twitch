package phrasebook

import (
	"fmt"
	"os"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"
)

type rawBook map[string]map[string][]rawPhrase

type rawPhrase struct {
	Text   string  `yaml:"text"`
	Weight float64 `yaml:"weight"`
}

type Phrasebook struct {
	entries map[string]map[string][]phrase
}

type phrase struct {
	Text   *template.Template
	Weight float64
}

func NewPhrasebook(path string) (*Phrasebook, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read phrasebook file: %w", err)
	}

	var raw rawBook
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to unmarshal phrasebook: %w", err)
	}

	return buildPhrasebook(raw)
}

func (p *Phrasebook) ValidateEntries(pairs ...string) error {
	var missingEntries []string

	for _, pair := range pairs {
		commandKey, scenarioKey, ok := strings.Cut(pair, ".")
		if !ok {
			missingEntries = append(missingEntries, pair)
			continue
		}

		command, ok := p.entries[commandKey]
		if !ok {
			missingEntries = append(missingEntries, pair)
			continue
		}

		_, ok = command[scenarioKey]
		if !ok {
			missingEntries = append(missingEntries, pair)
			continue
		}
	}

	if len(missingEntries) > 0 {
		return fmt.Errorf("missing required entries: %s", strings.Join(missingEntries, ", "))
	}

	return nil
}

func buildPhrasebook(raw rawBook) (*Phrasebook, error) {
	book := &Phrasebook{entries: make(map[string]map[string][]phrase)}

	for command, scenarios := range raw {
		validatedScenarios := make(map[string][]phrase, len(scenarios))

		for scenario, rawPhrases := range scenarios {
			validatedPhrases, err := validatePhrases(command, scenario, rawPhrases)
			if err != nil {
				return nil, fmt.Errorf("failed to validate scenario %q: %w", scenario, err)
			}

			validatedScenarios[scenario] = validatedPhrases
		}

		book.entries[command] = validatedScenarios
	}

	return book, nil
}

func validatePhrases(command, scenario string, rawPhrases []rawPhrase) ([]phrase, error) {
	if len(rawPhrases) == 0 {
		return nil, fmt.Errorf("%s.%s has no phrases", command, scenario)
	}

	validatedPhrases := make([]phrase, 0, len(rawPhrases))

	for i, rawPhrase := range rawPhrases {
		if strings.TrimSpace(rawPhrase.Text) == "" {
			return nil, fmt.Errorf("%s.%s has empty text", command, scenario)
		}

		if rawPhrase.Weight < 0 {
			return nil, fmt.Errorf("%s.%s must have non-negative weight", command, scenario)
		}

		t, err := template.New(fmt.Sprintf("%s.%s[%d]", command, scenario, i)).
			Option("missingkey=error").
			Parse(rawPhrase.Text)
		if err != nil {
			return nil, fmt.Errorf("%s.%s[%d]: invalid template: %w", command, scenario, i, err)
		}

		validatedPhrases = append(validatedPhrases, phrase{
			Text:   t,
			Weight: rawPhrase.Weight,
		})
	}

	return validatedPhrases, nil
}
