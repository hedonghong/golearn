package gotmd

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
	"strings"
	"testing"
)

func TestTrie(t *testing.T) {
	n := &node{}
	pattern1 := "/index"
	pattern1_1 := "/index/1"
	pattern2 := "/home"
	pattern2_1 := "/home/*"
	pattern2_2 := "/home/:get/2"
	parts1 := parsePattern(pattern1)
	parts1_1 := parsePattern(pattern1_1)
	parts2 := parsePattern(pattern2)
	parts2_1 := parsePattern(pattern2_1)
	parts2_2 := parsePattern(pattern2_2)
	n.insert(pattern1, parts1, 0)
	n.insert(pattern1_1, parts1_1, 0)
	n.insert(pattern2, parts2, 0)
	n.insert(pattern2_1, parts2_1, 0)
	n.insert(pattern2_2, parts2_2, 0)

	node := n.search(parts1_1, 0)
	fmt.Println(node)

	node = n.search(parts2_2, 0)
	fmt.Println(node)

	params := make(map[string]string)
	if node != nil {
		parts := parsePattern(node.pattern)
		for index, part := range parts {
			if part[0] == ':' {
				params[part[1:]] = parts2_2[index]
			}
			if part[0] == '*' && len(part) > 1 {
				params[part[1:]] = strings.Join(parts2_2[index:], "/")
				break
			}
		}
	}
}

func TestChi(t *testing.T) {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("welcome"))
	})
	http.ListenAndServe(":3000", r)
}
