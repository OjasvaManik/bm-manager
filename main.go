package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

func main() {
	const name = "bm-manager"
	log.SetPrefix(name + "\t")
	configDirectory, err := os.UserConfigDir()
	if err != nil {
		log.Fatal(err)
	}
	fullPath := filepath.Join(configDirectory, name, "bm.conf")

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		log.Println(err)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0700); err != nil {
			log.Fatalf("Error creating directory bm-manager: %v", err)
		}

		f, err := os.Create(fullPath)
		if err != nil {
			log.Fatalf("Error creating file bm.conf: %v", err)
		}
		defer func(f *os.File) {
			err := f.Close()
			if err != nil {
				log.Fatalf("Error closing file bm.conf: %v", err)
			}
		}(f)
		log.Println("Created bm-manager/bm.conf")
	}

	data, err := os.ReadFile(fullPath)
	if err != nil {
		log.Fatalf("Error reading file bm.conf: %v", err)
	}
	re := regexp.MustCompile(`(?m)^location\s*=\s*(.+)$`)
	locationMatches := re.FindStringSubmatch(string(data))
	if len(locationMatches) < 2 {
		log.Fatal("Error: 'location' key not found in bm.conf")
	}

	location := strings.TrimSpace(locationMatches[1])
	if location == "" {
		log.Fatal("Error reading file bm.conf: empty location")
	}

	bmPath := filepath.Join(location, "bookmarks.json")
	if _, err := os.Stat(bmPath); os.IsNotExist(err) {
		log.Println(err)
		if err := os.MkdirAll(location, 0700); err != nil {
			log.Fatalf(`Error creating directory "%s": %v`, location, err)
		}

		f, err := os.Create(bmPath)
		if err != nil {
			log.Fatalf("Error creating file bookmarks.json: %v", err)
		}
		defer func(f *os.File) {
			err := f.Close()
			if err != nil {
				log.Fatalf("Error closing file bookmarks.json: %v", err)
			}
		}(f)
		log.Println("Created bookmarks.json")
	}

	re = regexp.MustCompile(`(?m)^browser\s*=\s*(.+)$`)
	browserMatches := re.FindStringSubmatch(string(data))
	if len(browserMatches) < 2 {
		log.Fatal("Error: 'browser' key not found in bm.conf")
	}

	browser := strings.TrimSpace(browserMatches[1])
	if browser == "" {
		log.Fatal("Error reading file bm.conf: empty browser")
	}

	if len(os.Args) < 2 {
		handleList(bmPath)
		return
	}

	switch os.Args[1] {
	case "list":
		handleList(bmPath)
	case "add":
		if len(os.Args) < 4 {
			log.Fatal("Usage: add <name> <url> [aliases]")
		}
		handleAdd(bmPath, os.Args[2:])
	case "del":
		if len(os.Args) < 3 {
			log.Fatal("Usage: del <index>")
		}
		handleDelete(bmPath, os.Args[2])
	case "open":
		args := os.Args[2:]
		if len(args) < 1 {
			log.Fatal("Usage: open <index|alias> [-b <browser>]")
		}
		alias := ""
		targetBrowser := browser

		for i := 0; i < len(args); i++ {
			if args[i] == "-b" && i+1 < len(args) {
				targetBrowser = args[i+1]
				i++
			} else if !strings.HasPrefix(args[i], "-") && alias == "" {
				alias = args[i]
			}
		}
		if alias == "" {
			log.Fatal("Usage: open <index|alias> [-b <browser>]")
		}
		handleOpen(bmPath, targetBrowser, alias)
	default:
		handleList(bmPath)
	}

}

type Bookmark struct {
	Name  string   `json:"name"`
	URL   string   `json:"url"`
	Alias []string `json:"alias,omitempty"`
}

func handleList(bmPath string) {
	data, err := os.ReadFile(bmPath)
	if err != nil {
		log.Fatalf("Error reading file bookmarks.json: %v", err)
	}

	bookmarks := loadBookmarks(data)
	if len(bookmarks) == 0 {
		fmt.Println("Bookmarks is empty")
		return
	}

	for i, bm := range bookmarks {
		fmt.Printf("%d. \033]8;;%s\033\\%s\033]8;;\033\\\n", i+1, bm.URL, bm.Name)

		if len(bm.Alias) == 0 {
			fmt.Printf("   └── \033]8;;%s\033\\%s\033]8;;\033\\\n", bm.URL, bm.URL)
		} else {
			fmt.Printf("   └── \033]8;;%s\033\\%s\033]8;;\033\\\n", bm.URL, bm.URL)

			for j, alias := range bm.Alias {
				prefix := "       ├──"
				if j == len(bm.Alias)-1 {
					prefix = "       └──"
				}
				fmt.Printf("%s %s\n", prefix, alias)
			}
		}
	}
}

func handleAdd(bmPath string, args []string) {
	if len(args) < 2 || len(args) > 3 {
		log.Fatal("Usage: add <name> <url> [aliases]")
	}
	name := args[0]
	url := args[1]
	var aliases []string
	if len(args) == 3 && args[2] != "" {
		for _, a := range strings.Split(args[2], ",") {
			if trimmed := strings.TrimSpace(a); trimmed != "" {
				aliases = append(aliases, trimmed)
			}
		}
	}

	if name == "" || url == "" {
		log.Fatal("Fields cannot be empty string")
	}

	data, err := os.ReadFile(bmPath)
	if err != nil {
		log.Fatalf("Error reading file bookmarks.json: %v", err)
	}

	var bookmarks []Bookmark
	if strings.TrimSpace(string(data)) != "" {
		bookmarks = loadBookmarks(data)
	}

	bookmarks = append(bookmarks, Bookmark{Name: name, URL: url, Alias: aliases})
	saveBookmarks(bookmarks, bmPath)

	fmt.Println("Successfully Bookmarked")
}

func handleDelete(bmPath string, index string) {
	if index == "" {
		log.Fatalf("Index cannot be empty string")
	}

	data, err := os.ReadFile(bmPath)
	if err != nil {
		log.Fatalf("Error reading file bookmarks.json: %v", err)
	}

	if strings.TrimSpace(string(data)) == "" {
		fmt.Println("Bookmarks is empty, index doesn't exist")
		return
	}
	bookmarks := loadBookmarks(data)

	i, err := strconv.Atoi(index)
	if err != nil {
		log.Fatalf("Error converting index to int: %v", err)
	}
	i--

	if i < 0 || i >= len(bookmarks) {
		log.Fatalf("Index out of bounds: %d", i)
	}
	bookmarks = append(bookmarks[:i], bookmarks[i+1:]...)
	saveBookmarks(bookmarks, bmPath)
	fmt.Println("Successfully Deleted")
}

func handleOpen(bmPath string, browser string, arg string) {
	data, err := os.ReadFile(bmPath)
	if err != nil {
		log.Fatalf("Error reading file bookmarks.json: %v", err)
	}
	if strings.TrimSpace(string(data)) == "" {
		fmt.Println("Bookmarks is empty, index doesn't exist")
	}
	bookmarks := loadBookmarks(data)

	if i, err := strconv.Atoi(arg); err == nil {
		i--
		if i < 0 || i >= len(bookmarks) {
			log.Fatalf("Index out of bounds: %d", i+1)
		}
		openInBrowser(browser, bookmarks[i].URL, bookmarks[i].Name)
		return
	}

	for _, bm := range bookmarks {
		for _, alias := range bm.Alias {
			if alias == arg {
				openInBrowser(browser, bm.URL, bm.Name)
				return
			}
		}
	}

	log.Fatalf("No bookmark found for: %s", arg)
}

func loadBookmarks(data []byte) (bookmarks []Bookmark) {
	if err := json.Unmarshal(data, &bookmarks); err != nil {
		log.Fatalf("Invalid bookmarks.json format. Fix the JSON. Details: %v", err)
	}
	return bookmarks
}

func saveBookmarks(bookmarks []Bookmark, bmPath string) {
	data, err := json.MarshalIndent(bookmarks, "", "  ")
	if err != nil {
		log.Fatalf("Error encoding bookmarks.json: %v", err)
	}
	if err := os.WriteFile(bmPath, data, 0600); err != nil {
		log.Fatalf("Error writing bookmarks.json: %v", err)
	}
}

func openInBrowser(browser string, url string, name string) {
	cmd := exec.Command(browser, url)
	if err := cmd.Start(); err != nil {
		log.Fatalf("Error opening %s in browser: %v", name, err)
	}
	fmt.Printf("Opening %s in browser %s succeeded.\n", name, browser)
}
