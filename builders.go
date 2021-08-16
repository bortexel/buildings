package main

import (
	"compress/gzip"
	"encoding/binary"
	"encoding/csv"
	"errors"
	"fmt"
	"github.com/bortexel/buildings/litematica"
	"github.com/go-chi/chi/v5"
	"github.com/urfave/cli/v2"
	"gopkg.in/guregu/null.v4"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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
		router.Get("/", view("home", func(r *http.Request) interface{} { return nil }))

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

					ext := filepath.Ext(path)
					projectName := strings.TrimSuffix(filepath.Base(path), ext)
					items := make(map[string]uint)

					switch ext {
					case ".csv":
						reader := csv.NewReader(file)
						rows, err := reader.ReadAll()
						if err != nil {
							return err
						}

						itemNameIndex, itemAmountIndex := -1, -1

						for i, row := range rows {
							if i == 0 {
								for j, columnName := range row {
									if columnName == "Item" {
										itemNameIndex = j
									}

									if columnName == "Total" {
										itemAmountIndex = j
									}
								}

								if itemNameIndex < 0 || itemAmountIndex < 0 {
									return errors.New("unable to determine indexes of required columns: \"Item\" and \"Total\"")
								}

								continue
							}

							name := row[itemNameIndex]
							amountString := row[itemAmountIndex]
							amount, err := strconv.Atoi(amountString)
							if err != nil {
								return err
							}

							items[name] = uint(amount)
						}
					default:
						reader, err := gzip.NewReader(file)
						if err != nil {
							return err
						}

						liteProject, err := litematica.Load(reader)
						if err != nil {
							return err
						}

						if liteProject.Metadata.Name != "Unnamed" {
							projectName = liteProject.Metadata.Name
						}

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
									if int(b) >= len(region.BlockStatePalette) {
										continue
									}

									item := region.BlockStatePalette[b]
									items[item.Name] += 1
								}
							}
						}
					}

					project := &Project{
						Name:     projectName,
						Progress: 0,
					}

					Database.Save(project)
					log.Println("Successfully created project", project.Name, "with ID", project.ID)

					locale, err := LoadLocale(context.String("locale"))
					if err != nil {
						return err
					}

					for item, amount := range items {
						if item == "minecraft:air" {
							continue
						}

						name := strings.TrimSpace(item)
						if localizedName, ok := locale.Translations["block."+strings.ReplaceAll(name, ":", ".")]; ok {
							name = localizedName
						}

						if strings.ToLower(item) == item {
							// We have ID in "item"
							project.CreateResource(null.StringFrom(item), name, amount)
						} else {
							// We have item name in "item"
							project.CreateResource(null.String{}, name, amount)
						}
					}

					log.Println("Successfully created", len(items), "resources associated with project", project.Name)
					return nil
				},
			},
			{
				Name: "member",
				Subcommands: []*cli.Command{
					{
						Name: "create",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name: "name",
							},
						},
						Action: func(context *cli.Context) error {
							name := context.String("name")
							member := Member{
								Name: name,
							}
							Database.Save(&member)
							log.Println("Member with ID", member.ID, "has been successfully created")
							return nil
						},
					},
					{
						Name: "describe",
						Flags: []cli.Flag{
							&cli.Int64Flag{
								Name:     "id",
								Required: true,
							},
							&cli.BoolFlag{
								Name:    "clear",
								Aliases: []string{"c"},
							},
							&cli.StringFlag{
								Name: "note",
							},
						},
						Action: func(context *cli.Context) error {
							var member Member
							Database.Find(&member, context.Int64("id"))
							if !member.IsValid() {
								return errors.New("member not found")
							}

							if context.Bool("clear") {
								member.Note = null.String{}
								log.Println("Clearing note for member", member.Name)
							} else {
								member.Note = null.StringFrom(context.String("note"))
								log.Println("Setting note for member", member.Name, "to", member.Note)
							}

							Database.Save(member)
							return nil
						},
					},
				},
			},
			{
				Name: "project",
				Subcommands: []*cli.Command{
					{
						Name: "describe",
						Flags: []cli.Flag{
							&cli.Int64Flag{
								Name:     "id",
								Required: true,
							},
							&cli.BoolFlag{
								Name:    "clear",
								Aliases: []string{"c"},
							},
							&cli.StringFlag{
								Name: "description",
							},
						},
						Action: func(context *cli.Context) error {
							var project Project
							Database.Find(&project, context.Int64("id"))
							if !project.IsValid() {
								return errors.New("project not found")
							}

							if context.Bool("clear") {
								project.Description = null.String{}
								log.Println("Clearing description for project", project.Name)
							} else {
								project.Description = null.StringFrom(context.String("description"))
								log.Println("Setting description for project", project.Name, "to", project.Description)
							}

							Database.Save(project)
							return nil
						},
					},
					{
						Name: "rename",
						Flags: []cli.Flag{
							&cli.Int64Flag{
								Name:     "id",
								Required: true,
							},
							&cli.StringFlag{
								Name:     "name",
								Required: true,
							},
						},
						Action: func(context *cli.Context) error {
							var project Project
							Database.Find(&project, context.Int64("id"))
							if !project.IsValid() {
								return errors.New("project not found")
							}

							name := context.String("name")
							log.Println("Renaming project", project.Name, "to", name)
							project.Name = name
							Database.Save(project)
							return nil
						},
					},
				},
			},
		},
	}

	err = app.Run(os.Args)
	if err != nil {
		log.Fatalln(err)
	}
}
