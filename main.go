package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/rojters/opengraph"
	"gopkg.in/urfave/cli.v1"
)

// Artwork : Struct to store the relevant opengraph metadata for saving the image
type Artwork struct {
	Title    string
	ImageURL string
}

const defaultDir string = "~/Downloads"

func main() {
	var filename string
	var directory string
	var url string

	app := cli.NewApp()

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "filename, file, f",
			Usage:       "Save image as `FILE`",
			Value:       "",
			Destination: &filename,
		},
		cli.StringFlag{
			Name:        "directory, dir, d",
			Usage:       "Save image to `DIR`",
			Value:       "",
			Destination: &directory,
		},
	}

	app.Action = func(context *cli.Context) error {
		if context.NArg() > 0 {
			url = context.Args()[0]
		} else {
			cli.ShowAppHelp(context)
			return nil
		}

		if checkIsEmpty(directory) {
			directory = defaultDir
		}

		if directory == "." {
			wd, err := os.Getwd()
			if err != nil {
				log.Fatal("error loading directory")
			}

			directory = wd
		}

		destinationDir, err := homedir.Expand(directory)
		if err != nil {
			log.Fatal(err)
		}

		metaData := fetchMetaData(url)
		artwork := extractMetaData(metaData)

		if checkNotEmpty(artwork.Title) && checkNotEmpty(artwork.ImageURL) {
			if checkIsEmpty(filename) {
				filename = artwork.Title + ".jpg"
			}

			if strings.Contains(filename, "//") {
				filename = strings.Replace(filename, "//", "-", -1)
			}

			SaveFile(destinationDir, filename, artwork.ImageURL)

			fmt.Println("Downloaded " + filename + " to " + destinationDir)
		} else {
			fmt.Println("couldn't locate metadata")
		}

		return nil
	}

	app.Run(os.Args)
}

func fetchMetaData(url string) []opengraph.MetaData {
	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}

	md, err := opengraph.Extract(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	return md
}

func extractMetaData(md []opengraph.MetaData) Artwork {
	var artwork Artwork

	for i := range md {
		switch md[i].Property {
		case "title":
			artwork.Title = md[i].Content
		case "image":
			artwork.ImageURL = md[i].Content
		}
	}

	return artwork
}

// SaveFile - saves the image to a local file
func SaveFile(destDir string, filename string, imageURL string) {
	res, e := http.Get(imageURL)
	if e != nil {
		log.Fatal(e)
	}

	defer res.Body.Close()

	file, err := os.Create(path.Join(destDir, filename))
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	_, err = io.Copy(file, res.Body)
	if err != nil {
		log.Fatal(err)
	}
}

func checkIsEmpty(str string) bool {
	return (len(str) == 0)
}

func checkNotEmpty(str string) bool {
	return (len(str) > 0)
}
