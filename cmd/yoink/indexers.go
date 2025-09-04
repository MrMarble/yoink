package main

import (
	"fmt"

	yoink "github.com/mrmarble/yoink"
)

type IndexersCmd struct{}

func (i *IndexersCmd) Run(ctx *Context) error {
	manager, err := yoink.GetIndexerManager(ctx.config)
	if err != nil {
		return err
	}

	indexers, err := manager.GetIndexers()
	if err != nil {
		return err
	}

	// Display format depends on indexer type
	if ctx.config.IndexerType == "jackett" {
		fmt.Printf("ID          | Name\n----------- | ----\n")
		for _, indexer := range indexers {
			if indexer.IsEnabled() {
				fmt.Printf("%-11s | %s\n", indexer.GetID(), indexer.GetName())
			}
		}
	} else {
		fmt.Printf("ID   | Name\n---- | ----\n")
		for _, indexer := range indexers {
			if indexer.IsEnabled() {
				fmt.Printf("%-4d | %s\n", indexer.GetID(), indexer.GetName())
			}
		}
	}

	return nil
}
