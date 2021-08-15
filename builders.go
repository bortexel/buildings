package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

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

	err = Database.AutoMigrate(&Project{}, &Resource{})
	if err != nil {
		log.Fatalln("Unable to run migrations:", err)
		return
	}

	router := chi.NewRouter()
	router.Get("/projects/{id}", func(writer http.ResponseWriter, request *http.Request) {
		id := chi.URLParam(request, "id")
		project := &Project{}
		Database.Find(&project, id)

		body, err := json.Marshal(project)
		if err != nil {
			return
		}

		_, _ = writer.Write(body)
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
		},
	}

	err = app.Run(os.Args)
	if err != nil {
		log.Fatalln(err)
	}
}
