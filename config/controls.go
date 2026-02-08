package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type PlayerControls struct {
	Left  string `json:"left"`
	Right string `json:"right"`
	Jump  string `json:"jump"`
}

type Controls struct {
	Quit    string         `json:"quit"`
	Player1 PlayerControls `json:"player1"`
	Player2 PlayerControls `json:"player2"`
}

func (c *Controls) Normalize() {
	c.Quit = strings.ToUpper(c.Quit)
	c.Player1.Left = strings.ToUpper(c.Player1.Left)
	c.Player1.Right = strings.ToUpper(c.Player1.Right)
	c.Player1.Jump = strings.ToUpper(c.Player1.Jump)
	c.Player2.Left = strings.ToUpper(c.Player2.Left)
	c.Player2.Right = strings.ToUpper(c.Player2.Right)
	c.Player2.Jump = strings.ToUpper(c.Player2.Jump)
}

func LoadControls(path string) (Controls, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Controls{}, err
	}

	var c Controls
	if err := json.Unmarshal(b, &c); err != nil {
		return Controls{}, fmt.Errorf("parse %s: %w", path, err)
	}
	c.Normalize()
	return c, nil
}
