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
)

const defaultDir string = "~/Downloads"

func main() {
	args := os.Args[1:]

	destinationDir, err := homedir.Expand(defaultDir)
	if err != nil {
		log.Fatal(err)
	}

	if len(args) <= 0 {
		Usage()
		return
	}

	url := args[0]
	if url == "" {
		Usage()
		return
	}

	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}

	md, err := opengraph.Extract(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	var title, imageURL string

	for i := range md {
		switch md[i].Property {
		case "title":
			title = md[i].Content
		case "image":
			imageURL = md[i].Content
		}
	}

	if len(title) > 0 && len(imageURL) > 0 {
		title = title + ".jpg"

		if strings.Contains(title, "//") {
			title = strings.Replace(title, "//", "-", -1)
		}

		SaveFile(destinationDir, title, imageURL)

		fmt.Println("Downloaded " + title + " to " + destinationDir)
	} else {
		fmt.Println("couldn't locate metadata")
	}
}

// SaveFile - saves the image to a local file
func SaveFile(destDir string, title string, imageURL string) {
	res, e := http.Get(imageURL)
	if e != nil {
		log.Fatal(e)
	}

	defer res.Body.Close()

	file, err := os.Create(path.Join(destDir, title))
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	_, err = io.Copy(file, res.Body)
	if err != nil {
		log.Fatal(err)
	}
}

// Usage - Display usage info
func Usage() {
	fmt.Println("Please pass a url to check")
	fmt.Println("")
	fmt.Println("usage: scar https://soundcloud.com/artist/track")
	return
}
