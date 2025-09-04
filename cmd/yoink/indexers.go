package main

import (
	"fmt"

	"github.com/mrmarble/yoink/pkg/jackett"
	"github.com/mrmarble/yoink/pkg/prowlarr"
)

type IndexersCmd struct{}

func (i *IndexersCmd) Run(ctx *Context) error {
	if ctx.config.IndexerType == "jackett" {
		return i.runJackett(ctx)
	}
	return i.runProwlarr(ctx)
}

func (i *IndexersCmd) runProwlarr(ctx *Context) error {
	client := prowlarr.NewClient(ctx.config.Prowlarr.Host, ctx.config.Prowlarr.APIKey)
	indexers, err := client.GetIndexers()
	if err != nil {
		return err
	}

	fmt.Printf("ID   | Name\n---- | ----\n")
	for _, v := range indexers {
		if v.Enable && v.Protocol == "torrent" {
			fmt.Printf("%-4d | %s\n", v.ID, v.Name)
		}
	}

	return nil
}

func (i *IndexersCmd) runJackett(ctx *Context) error {
	client := jackett.NewClient(ctx.config.Jackett.Host, ctx.config.Jackett.APIKey)
	indexers, err := client.GetIndexers()
	if err != nil {
		return err
	}

	fmt.Printf("ID          | Name\n----------- | ----\n")
	for _, v := range indexers {
		fmt.Printf("%-11s | %s\n", v.ID, v.Title)
	}

	return nil
}
