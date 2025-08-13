package main

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"
)

func (benchmark *Benchmark) Json() (string, error) {
	prettyJSON, err := json.MarshalIndent(benchmark, "", "    ")
	if err != nil {
		return "", fmt.Errorf("error marshalling JSON: %w", err)
	}

	return string(prettyJSON), nil
}

func (benchmark *Benchmark) Yaml() (string, error) {
	yamlData, err := yaml.Marshal(&benchmark)
	if err != nil {
		return "", fmt.Errorf("error marshalling yaml: %v", err)
	}

	return string(yamlData), nil
}
