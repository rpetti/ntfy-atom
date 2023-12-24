package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gorilla/feeds"
	"github.com/gorilla/mux"
)

var (
	ntfy_url *url.URL
)

func getHealthCheck(w http.ResponseWriter, r *http.Request) {
	_, err := io.WriteString(w, "OK!")
	if err != nil {
		log.Printf("error sending healthcheck to client: %v", err)
	}
}

func getTopic(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	topic := vars["topic"]
	since := r.URL.Query().Get("since")
	if since != "" {
		since = "7d"
	}
	feed, err := feedify(topic, since)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, err2 := io.WriteString(w, err.Error())
		if err2 != nil {
			log.Printf("error sending feed error back to client: %v", err2)
		}
		log.Printf("error fetching topic %s: %v", topic, err)
		return
	}
	_, err = io.WriteString(w, feed)
	if err != nil {
		log.Printf("error sending feed back to client: %v", err)
	}
}

type NtfyEvent struct {
	Tags     []string `json:"tags"`
	Event    string   `json:"event"`
	Topic    string   `json:"topic"`
	Priority int      `json:"priority"`
	Title    string   `json:"title"`
	Message  string   `json:"message"`
	Time     int64    `json:"time"`
}

// feedify
// fetch feed from ntfy, and translate it to an atom feed
func feedify(topic string, since string) (string, error) {
	reqUrl := ntfy_url.JoinPath(topic, "json")
	q := reqUrl.Query()
	q.Add("poll", "1")
	q.Add("since", since)
	reqUrl.RawQuery = q.Encode()
	resp, err := http.Get(reqUrl.String())
	if err != nil {
		return "", err
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

    log.Println(string(b))
    feed_items := []*feeds.Item{}

	for _, raw_event := range strings.Split(string(b), "\n") {
        if strings.Trim(raw_event, " ") == "" {
            continue
        }
        event := &NtfyEvent{}
        err := json.Unmarshal([]byte(raw_event), event)
        if err != nil {
            log.Printf("failed to parse %s", raw_event)
            return "", err
        }

        feed_items = append(feed_items, &feeds.Item{
            Title: event.Title,
            Content: event.Message,
            Created: time.Unix(event.Time, 0),
        })
	}
    feed := &feeds.Feed{
        Title: fmt.Sprintf("ntfy topic %s", topic),
        Link: &feeds.Link{Href: ntfy_url.JoinPath(topic).String()},
        Items: feed_items,
    }

	return feed.ToAtom()
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/health", getHealthCheck)
	r.HandleFunc("/topics/{topic}", getTopic)

	port := os.Getenv("NTFY_ATOM_PORT")
	if port == "" {
		port = "8080"
	}

	if os.Getenv("NTFY_URL") == "" {
		log.Fatalf("NTFY_URL not set, cannot start")
	}
	var err error
	ntfy_url, err = url.Parse(os.Getenv("NTFY_URL"))
	if err != nil {
		log.Fatalf("not a valid ntfy url: %v", err)
	}

	log.Println("listen on", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), r))
}
