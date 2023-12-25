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

	"github.com/google/uuid"
	"github.com/gorilla/feeds"
	"github.com/gorilla/mux"
)

var (
	ntfy_url *url.URL
    uuid_namespace uuid.UUID
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
	if since == "" {
		since = "168h"
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

    feed_items := []*feeds.Item{}


    updated := time.Unix(0,0)

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
        event_time := time.Unix(event.Time, 0)
        if event_time.After(updated) {
            updated = event_time
        }

        event_uuid := uuid.NewSHA1(uuid_namespace, []byte(fmt.Sprintf(
            "%s:%s:%d",
            event.Title,
            event.Message,
            event.Time,
        )))

        feed_items = append(feed_items, &feeds.Item{
            Title: event.Title,
            Content: event.Message,
            Created: event_time,
            Author: &feeds.Author{Name: "ntfy", Email: fmt.Sprintf("ntfy@%s", ntfy_url.Hostname())},
            Id: fmt.Sprintf("urn:uuid:%s", event_uuid.String()),
        })
	}
    feed := &feeds.Feed{
        Title: fmt.Sprintf("ntfy topic %s", topic),
        Link: &feeds.Link{Href: ntfy_url.JoinPath(topic).String()},
        Items: feed_items,
        Updated: updated,
        Author: &feeds.Author{Name: "ntfy", Email: fmt.Sprintf("ntfy@%s", ntfy_url.Hostname())},
    }
	return feed.ToAtom()
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/health", getHealthCheck)
	r.HandleFunc("/topics/{topic}", getTopic)

    var err error

    uuid_namespace, err = uuid.Parse("97ef7f2e-9733-4bf3-ac69-4ba1c59ca656")
    if err != nil {
        log.Fatalf("%v", err)
    }

	port := os.Getenv("NTFY_ATOM_PORT")
	if port == "" {
		port = "8080"
	}

	if os.Getenv("NTFY_URL") == "" {
		log.Fatalf("NTFY_URL not set, cannot start")
	}

	ntfy_url, err = url.Parse(os.Getenv("NTFY_URL"))
	if err != nil {
		log.Fatalf("not a valid ntfy url: %v", err)
	}

	log.Println("listen on", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), r))
}
