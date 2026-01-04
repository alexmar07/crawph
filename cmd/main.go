package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/alexmar07/crawler-go/crawph"
)

func main() {

	var urls string

	flag.StringVar(&urls, "urls", "", "Lists of urls for Crawph")

	flag.Parse()

	listUrls := strings.Split(urls, ",")

	ch := crawph.NewCrawph(5)

	if len(listUrls) == 0 {
		panic("No urls")
	}

	fmt.Println("Crawph is running...")

	ch.Start(listUrls)

	fmt.Println("Crawph stopped...")

}
