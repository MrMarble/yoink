package main

import (
	"encoding/json"
	"fmt"
)

type PrintConfigCmd struct{}

func (p *PrintConfigCmd) Run(ctx *Context) error {
	parsedConfig, err := json.MarshalIndent(ctx.config, "", "  ")
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", string(parsedConfig))
	return nil
}
