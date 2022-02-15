package main

import (
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"
)

// https://raw.githubusercontent.com/getbread/financing/master/go.sum?token=XXXXXXXX
// https://raw.githubusercontent.com/getbread/workflow/master/service/go.sum?token=XXXXXXXX
// https://raw.githubusercontent.com/getbread/featureflag/master/service/go.sum?token=XXXXXXXX
// https://raw.githubusercontent.com/getbread/gokit/master/go.sum?token=XXXXXXXX

// https://raw.githubusercontent.com/getbread/gokit/master/go.sum?token=AVRIM3HIMD2YY7UQ636GKZ3B2NAKQ
// https://raw.githubusercontent.com/getbread/gokit/master/go.sum?token=AVRIM3G5EWCCAWD5Y7HGADLB2NBR4

const (
	branch = "master"
	company = "getbread"
	urlPattern = "https://raw.githubusercontent.com/%s/%s/%s%s/go.sum?token=%s"
	local = "local"
)

var modules = []Module {
	{name: "gokit", path: ""},
	{name: "featureflag", path: "/service"},
	{name: "workflow", path: "/service"},
}

var slashReplacer = strings.NewReplacer("/", ".")

type Module struct {
	name string
	path string
}

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.SetPrefix("XYZ ")

	offline := flag.Bool("offline", false, "Work offline")
	user  := flag.String("user", "anonymous", "Github user")
	token := flag.String("token", "XXXXXXX", "Github token")
	flag.Parse()

	if *offline {
		log.Println("Working offline")
	} else {
		download (*user, *token)
	}
}


func download(user string, token string) {
	cleanup()

	log.Printf("Dowloading %d modules\n", len(modules))
	for _, mod := range modules {
		downloadMod(mod, user, token)
	}
}

func cleanup() {
	log.Println("Cleaning up files")
	fileInfo, err := os.Stat(local)
	if err == nil {
		if fileInfo.IsDir() {
			log.Println("File does exist. File information:")
			dir, err := os.Open(local)
			if err != nil {
				log.Fatalf("Faile to open directory %s, %v", local, err)
			}
			files, err := dir.Readdir(0)
			if err != nil {
				log.Fatalf("Faile to read files from %s, %v", local, err)
			}
			for _, file := range files {
				err = os.Remove(local + "/" + file.Name())
				if err != nil {
					log.Fatalf("Faile to remove file %s, %v", file.Name(), err)
				}
			}
			log.Printf("Removed %d files from %s", len(files), local)
		} else {
			log.Fatalf("%s is not a directory", local)
		}
	} else {
		if os.IsNotExist(err) {
			log.Println("Directory does not exist, creating")
			if createError := os.Mkdir(local, fs.ModeDir | 0700); createError != nil {
				log.Fatalf("can not create %s directory, %v", local, createError)
			}
		} else {
			log.Fatalf("can not open %s directory, %v", local, err)
		}
	}
}

func downloadMod(mod Module, username string, token string) {
	// Evaluate the URL
	url := fmt.Sprintf(urlPattern, company, mod.name, branch, mod.path, token)

	req, err := http.NewRequest(http.MethodGet, url, http.NoBody)
	if err != nil {
		log.Fatalf("failed to create request for %s, error %d", url, err)
	}
	// GitHub Auth
	req.SetBasicAuth(username, token)

	// Get the data
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("failed to download %s, %v", url, err)
	}
	defer closer(resp.Body)

	// Check response code
	if resp.StatusCode != 200 {
		log.Fatalf("failed to download %s, response code %d", url, resp.StatusCode)
	}

	// Evaluate the file name
	middlePath := ""
	if mod.path != "" {
		middlePath = slashReplacer.Replace(mod.path)
	}
	fileName := local + "/" + mod.name + middlePath + ".go.sum"
	// Create the file
	out, err := os.Create(fileName)
	if err != nil {
		log.Fatalf("failed to create a file %s for URL %s, error: %v", fileName, url, err)
	}
	defer closer(out)

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		log.Fatalf("failed to copy a URL %s to file %s, error: %v", fileName, url, err)
	}
	log.Printf("Downloaded %s -> %s\n", url, fileName)
}

func closer(closer io.Closer) {
	err := closer.Close()
	if err != nil {
		log.Println("Opps, close error", err)
	}
}
