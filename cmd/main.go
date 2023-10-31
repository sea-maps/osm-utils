package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"

	"github.com/sea-maps/osm-utils/internal/pbf"
	"github.com/sea-maps/osm-utils/internal/valhalla"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name: "osm-utils",
		Commands: []*cli.Command{
			{
				Name: "extract-objects",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "pbf-file",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "type",
						Required: true,
						Action: func(c *cli.Context, s string) error {
							if s != "node" && s != "way" && s != "relation" {
								return fmt.Errorf("type must be node, way or relation")
							}
							return nil
						},
					},
					&cli.StringSliceFlag{
						Name: "refs",
					},
					&cli.StringSliceFlag{
						Name: "tags",
					},
				},
				Action: func(c *cli.Context) error {
					return pbf.ExtractObject(c.Context,
						mustOpenReadFile(c.String("pbf-file")),
						c.String("type"),
						c.StringSlice("refs"),
						c.StringSlice("tags"),
					)
				},
			},
			{
				Name: "compare-batch-routes",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "route-dir",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "valhalla-url",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "costing-json",
						Required: true,
					},
				},
				Action: func(c *cli.Context) error {
					return valhalla.CompareBatchRoutes(c.Context,
						os.DirFS(c.String("route-dir")),
						c.String("valhalla-url"),
						mustOpenReadFile(c.String("costing-json")),
					)
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func mustOpenReadFile(fileName string) fs.File {
	f, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}
	return f
}
