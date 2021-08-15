package main

import (
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/bortexel/buildings/litematica"
	"github.com/go-chi/chi/v5"
	"github.com/urfave/cli/v2"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	Database *gorm.DB

	sqlHost     = os.Getenv("BB_SQL_HOST")
	sqlUser     = os.Getenv("BB_SQL_USER")
	sqlPassword = os.Getenv("BB_SQL_PASSWORD")
	sqlDatabase = os.Getenv("BB_SQL_DATABASE")
)

func main() {
	var err error
	Database, err = gorm.Open(mysql.Open(fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?parseTime=true&loc=%s", sqlUser, sqlPassword, sqlHost, sqlDatabase, "Europe%2FMoscow")))
	if err != nil {
		log.Fatalln("Unable to connect to database:", err)
		return
	}

	err = Database.AutoMigrate(&Project{}, &Resource{}, &Member{})
	if err != nil {
		log.Fatalln("Unable to run migrations:", err)
		return
	}

	router := chi.NewRouter()
	router.Route("/", func(r chi.Router) {
		router.Get("/projects", view("projects", func(r *http.Request) interface{} {
			return AllProjects()
		}))

		router.Get("/projects/{id}", view("project", func(r *http.Request) interface{} {
			return ProjectPage(r)
		}))

		router.Get("/members", view("members", func(r *http.Request) interface{} {
			return AllMembers()
		}))
	})

	router.Route("/api/v1", func(router chi.Router) {
		router.Get("/projects", endpoint(ListProjects))
		router.Get("/projects/{id}", endpoint(FindProject))

		router.Get("/resources", endpoint(ListResources))
	})

	app := cli.App{
		Commands: []*cli.Command{
			{
				Name:        "serve",
				Description: "Starts webserver",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "addr",
						Value: ":8080",
					},
				},
				Action: func(context *cli.Context) error {
					addr := context.String("addr")
					log.Println("Starting webserver on", addr)
					return http.ListenAndServe(addr, router)
				},
			},
			{
				Name:        "import",
				Description: "Imports litematica project",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "file",
						Aliases: []string{"f"},
						Value:   "project.litematica",
					},
					&cli.StringFlag{
						Name:    "locale",
						Aliases: []string{"l"},
						Value:   "data/locale.json",
					},
				},
				Action: func(context *cli.Context) error {
					path := context.String("file")
					file, err := os.Open(path)
					if err != nil {
						return err
					}

					defer file.Close()
					reader, err := gzip.NewReader(file)
					if err != nil {
						return err
					}

					liteProject, err := litematica.Load(reader)
					if err != nil {
						return err
					}

					project := &Project{
						Name:     liteProject.Metadata.Name,
						Progress: 0,
					}

					Database.Save(project)

					items := make(map[string]uint)
					for name, region := range liteProject.Regions {
						log.Println("Processing region", name)
						for _, state := range region.BlockStates {
							if state == 0 {
								continue
							}

							// https://github.com/maruohon/litematica/issues/53#issuecomment-520281566
							// IDK why 256, let's just hope that it's enough
							bytes := make([]byte, 256)
							binary.BigEndian.PutUint64(bytes[:], uint64(state))
							for _, b := range bytes {
								if int(b) > len(region.BlockStatePalette) {
									continue
								}

								item := region.BlockStatePalette[b]
								items[item.Name] += 1
							}
						}
					}

					locale, err := LoadLocale(context.String("locale"))
					if err != nil {
						return err
					}

					for item, amount := range items {
						if item == "minecraft:air" {
							continue
						}

						name := item
						if localizedName, ok := locale.Translations["block."+strings.ReplaceAll(name, ":", ".")]; ok {
							name = localizedName
						}

						project.CreateResource(item, name, amount)
					}

					return nil
				},
			},
		},
	}

	err = app.Run(os.Args)
	if err != nil {
		log.Fatalln(err)
	}
}
