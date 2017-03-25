package server

import (
	"errors"
	"net/url"
	"path"
	"strings"

	"github.com/henrylee2cn/rpc2/common"
)

// URIFormator URI format tool
type URIFormator interface {
	// URIEncode encode the parmaters to URI.
	URIEncode(query url.Values, pathSegment ...string) (uri string)
	// URIParse parses URI and returns parmaters(path and query).
	URIParse(uri string) (path string, query url.Values, err error)
}

// URLFormat implements a URIFormator in URL format.
type URLFormat struct{}

// URIEncode encode the parmaters to uri.
func (u *URLFormat) URIEncode(query url.Values, pathSegment ...string) (uri string) {
	for i := len(pathSegment) - 1; i >= 0; i-- {
		pathSegment[i] = common.SnakeString(pathSegment[i])
	}
	sm := url.URL{
		Path: path.Join(pathSegment...),
	}
	if !strings.HasPrefix(sm.Path, "/") {
		sm.Path = "/" + sm.Path
	}
	if len(query) != 0 {
		sm.RawQuery = query.Encode()
	}
	return sm.String()
}

// URIParse parses URI and returns parmaters(path and query).
func (u *URLFormat) URIParse(uri string) (path string, query url.Values, err error) {
	if uri == "" {
		return
	}
	sm, err := url.Parse(uri)
	if err != nil {
		err = errors.New("The URI '" + uri + "' is not formatted correctly")
		return
	}
	return sm.Path, sm.Query(), nil
}
