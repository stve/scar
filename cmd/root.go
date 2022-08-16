package cmd

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/johnreutersward/opengraph"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)

var filename string
var directory string

var rootCmd = &cobra.Command{
	Use:   "scar https://soundcloud.com/artist/track [flags]",
	Short: "scar is a simple cli for downloading soundcloud artwork",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		download(args[0], filename, directory)
	},
}

func init() {
	rootCmd.Flags().StringVarP(&filename, "filename", "", "", "Specify the filename of the image")
	rootCmd.Flags().StringVarP(&directory, "directory", "", "", "Specify the directory to save the image")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// Artwork : Struct to store the relevant opengraph metadata for saving the image
type Artwork struct {
	Title    string
	ImageURL string
}

const defaultDir string = "~/Downloads"

func download(url string, filename string, directory string) {
	destinationDir, err := resolveDestination(directory)
	if err != nil {
		log.Fatal(err)
	}

	metaData, err := fetchMetaData(url)
	if err != nil {
		log.Fatal(err)
	}

	artwork := extractMetaData(metaData)

	if artwork.Title != "" && artwork.ImageURL != "" {
		if filename == "" {
			filename = artwork.Title + ".jpg"
		}

		if strings.Contains(filename, "/") {
			filename = strings.Replace(filename, "/", "-", -1)
		}

		saveFile(destinationDir, filename, artwork.ImageURL)

		fmt.Println("Downloaded " + filename + " to " + destinationDir)
	} else {
		fmt.Println("couldn't locate metadata")
	}
}

func resolveDestination(directory string) (string, error) {
	// if empty directory, use default
	if directory == "" {
		directory, err := homedir.Expand(defaultDir)
		if err != nil {
			return "", err
		}

		return directory, nil
	}

	// if local working directory, resolve it
	if directory == "." {
		wd, err := os.Getwd()
		if err != nil {
			return "", err
		}

		return wd, nil
	}

	// if anything else, resolve it
	dir, err := filepath.Abs(directory)
	if err != nil {
		return "", err
	}

	return dir, nil
}

func fetchMetaData(url string) ([]opengraph.MetaData, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	md, err := opengraph.Extract(res.Body)
	if err != nil {
		return nil, err
	}

	return md, nil
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
func saveFile(destDir string, filename string, imageURL string) {
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
