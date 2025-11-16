package download

import (
	"io"
	"log"
	"net/http"
)

func DownloadFromURL(url string) string {
	log.Println("downloading file: " + url)
	resp, err := http.Get(url)
	if err != nil {
		log.Println(err)
		return ""
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return ""
	}

	return string(body)
}
