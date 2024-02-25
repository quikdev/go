// greeting.go
package greeting

import (
	"fmt"
	"net/http"
)

func Format(w http.ResponseWriter, name, version, description, buildTime string) {
	// Set the content type header
	w.Header().Set("Content-Type", "text/plain")

	// Write server information to the response writer
	fmt.Fprintf(w, "Server: %s v%s\n%s\nCreated on %s\n", name, version, description, buildTime)
}
