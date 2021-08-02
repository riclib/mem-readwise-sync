# mem-readwise-sync
A command line synchronizer for bringing your books, articles and highlights from readwise to mem. 
Every time you run it it will only bring the new highlights.

Books will be stored separately from highlights. When there are new highlights a new mem will be created with just the new hghlights.

```
Book <- Initial hilights
     <- Highlights you took a few days later
     <- And some more
```
You can find highlights for the synchronized books as related information, as they link to the book.

## Getting Started

Download the release for your platform [here](https://github.com/riclib/mem-readwise-sync/releases). After extracting you will have three files:

```
mem-readwise-sync
config.yml
README.md
```

[Create a readwise access token](https://readwise.io/access_token) and [mem api token](https://mem.ai/flows/api) and add them to the config.yml file with your favorite text editor.

``` yaml
# Replace enter_your_token_here by readwise token you can generate at https://readwise.io/access_token
readwise_key: enter_your_token_here

# replace enter_your_key_here by mem api key you can generate at https://mem.ai/flows/api or by clicking Flows than Api - configure
mem_key: enter_your_key_here
```

## Running mem-readwise-sync

```bash
usage: mem-readwise-sync [<flags>]

Flags:
  -h, --help                     Show context-sensitive help (also try
                                 --help-long and --help-man).
      --config.file=CONFIG.FILE  Configuration file.
      --version                  Show application version.
```

after editing the `config.yml` file you can run with the following command:

``` bash
./mem-readwise-sync --config.file=config.yml
```

after the first run `mem-readwise-sync` will create a database file in the current folder called `mem-readwise-sync.db`

This file contains the id's of the already synced books, as well as a cache of all your readwise books to speed up running. It also stores the time of the last run, so next time you run you will only get the new highlights.

## Changing the templates

You can change the Templates to create books and highlights. They are in this section of the config.yml file:

```yaml
## You can change the template for book and for highlight below. Make sure you keep the 2 leading spaces on each line and the two blank lines between. See README.md or the github repo for instructions and valid fields

book_template: |
  # {{.Title}}
  #{{.Category}} #readwise
  Author: {{.Author}}
  ![Cover]({{.CoverImageUrl}})
  {{if .SourceUrl}} src: {{.SourceUrl}} {{end}}


highlight_template: |
  # {{.Book.MemURL}} highlights
  #highlights synced on {{.TimeStamp}}
  {{range .Highlight}}- {{.Text}}{{range .Tags}} #{{.Name}}{{end}}{{if .Note}}
    - {{.Note}}{{end}}
  {{end}}
```

Add any field you want with {{ .FieldName }}. To understand how  to change go templates refer to [Using Go Templates](https://blog.gopheracademy.com/advent-2017/using-go-templates/)

Fields available are:

``` go
type Book struct {
	Id              int           `json:"id"`
	Title           string        `json:"title"`
	Author          string        `json:"author"`
	Category        string        `json:"category"`
	NumHighlights   int           `json:"num_highlights"`
	LastHighlightAt time.Time     `json:"last_highlight_at"`
	Updated         time.Time     `json:"updated"`
	CoverImageUrl   string        `json:"cover_image_url"`
	HighlightsUrl   string        `json:"highlights_url"`
	SourceUrl       interface{}   `json:"source_url"`
	Asin            string        `json:"asin"`
	Tags            []interface{} `json:"tags"`
	MemURL          string        `json:"mem_url"`
}
```

and

``` go
type Highlight struct {
	Id            int       `json:"id"`
	Text          string    `json:"text"`
	Note          string    `json:"note"`
	Location      int       `json:"location"`
	LocationType  string    `json:"location_type"`
	HighlightedAt time.Time `json:"highlighted_at"`
	Url           string    `json:"url"`
	Color         string    `json:"color"`
	Updated       time.Time `json:"updated"`
	BookId        int       `json:"book_id"`
	Tags          []struct {
		Id   int    `json:"id"`
		Name string `json:"name"`
	} `json:"tags"`
}
```

