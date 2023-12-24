# ntfy-atom

A microservice for hosting atom feeds from ntfy topics.

## Configuration

Configuration is set entirely via environment variables:

- `NTFY_URL` - ntfy base url
- `NTFY_ATOM_PORT` - port number to host service on (optional, 8080 default)

Note: Does not currently support authentication or protected topics.

## Use

Point your feedreader at the ntfy-atom host:port, with the topic in the URI:

Feed of notifications for `my-topic`:

`http://ntfy-atom:8080/topics/my-topic`

Feed of notifications for all topics:

`http://ntfy-atom:8080/all-topics`

Health check URL:

`http://ntfy-atom:8080/health`

