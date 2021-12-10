package main

import (
	"github.com/boltdb/bolt"
	"gopkg.in/alecthomas/kingpin.v2"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"text/template"
	"time"
)

const (
	lastSyncKey = "last_sync"
)

var (
	configFile = kingpin.Flag("config.file", "Configuration file.").ExistingFile()
	mem        = kingpin.Command("mem", "sync with mem.")
	craft      = kingpin.Command("craft", "Generate xml for craft shortcut")
)

type Config struct {
	ReadwiseKey       string `yaml:"readwise_key"`
	MemKey            string `yaml:"mem_key"`
	TimestampFormat   string `yaml:"timestamp_format"`
	BookTemplate      string `yaml:"book_template"`
	HighlightTemplate string `yaml:"highlight_template"`
}

type HighlightsOfBook struct {
	TimeStamp string
	Book      Book
	Highlight []Highlight
}

var (
	bookBucket       = []byte("Books")
	lastUpdateBucket = []byte("LastUpdate")
)

type Context struct {
	db           *bolt.DB
	config       Config
	templates    map[string]*template.Template
	lastSyncTime string
	thisSyncTime string
}

func main() {
	kingpin.Version("mem-readwise-sync 1.0.0")
	kingpin.HelpFlag.Short('h')
	command := kingpin.Parse()

	var (
		context Context
		err     error
	)

	context.config, err = getConfig()
	context.db, err = bolt.Open("mem-readwise-sync.db", 0644, nil)
	if err != nil {
		log.Panic("Opening db", err)
	}
	context.templates = getTemplates(context.config)
	context.thisSyncTime = time.Now().UTC().Format(time.RFC3339)

	switch command {
	case "mem":
		syncMem(context)

	}

	AddTimeToCache(context, context.thisSyncTime)
}

func getConfig() (Config, error) {

	var config Config
	if *configFile == "" {
		kingpin.Usage()
		os.Exit(1)
	}

	data, err := ioutil.ReadFile(*configFile)
	if err != nil {
		log.Panic("Error reading config", err)
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		log.Panic("Error parsing config", err)
	}
	return config, err
}

func getTemplates(config Config) map[string]*template.Template {
	templates := make(map[string]*template.Template, 2)
	templates["book"] = template.Must(template.New("book").Parse(config.BookTemplate))
	templates["highlight"] = template.Must(template.New("highlight").Parse(config.HighlightTemplate))
	return templates
}
