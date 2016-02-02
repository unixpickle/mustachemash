package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/unixpickle/mustachemash/mustacher"
)

func main() {
	mirrorFlag := flag.Bool("mirror", true, "mirror templates")
	includeNegatives := flag.Bool("negatives", false, "include negatives")
	strictness := flag.Float64("strictness", 0.2, "threshold strictness")
	userInfoPrefix := flag.String("uiprefix", "", "template userinfo prefix")

	flag.Parse()

	if len(flag.Args()) != 2 {
		fmt.Fprintln(os.Stderr, "Usage: encode-db [flags] <database-dir> <output.json>")
		flag.PrintDefaults()
		os.Exit(1)
	}

	log.Println("Loading database...")
	db, err := mustacher.LoadDatabase(flag.Args()[0], *strictness)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if *mirrorFlag {
		log.Println("Adding mirrors...")
		db.AddMirrors()
	}

	if !*includeNegatives {
		db.Negatives = []*mustacher.Image{}
	}

	for _, template := range db.Templates {
		template.UserInfo = (*userInfoPrefix) + template.UserInfo
	}

	log.Println("Encoding as JSON...")
	data, err := json.Marshal(db)
	if err != nil {
		panic(err)
	}

	log.Println("Saving...")
	if err := ioutil.WriteFile(flag.Args()[1], data, 0755); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
