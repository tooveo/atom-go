package atom

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

type serverResponse struct {
	statusCode int
	body       string
}

func Server(resp ...*serverResponse) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(resp[0].statusCode)
		io.WriteString(w, resp[0].body)
		resp = resp[1:]
	}))
}

func TestPost(t *testing.T) {
	responses := []*serverResponse{
		{http.StatusOK, "OK"},
		{http.StatusUnauthorized, "Unauthorized"},
		{http.StatusInternalServerError, "StatusInternalServerError"},
	}
	server := Server(responses...)
	base, _ := url.Parse(server.URL)
	c := &client{base}
	for i := range responses {
		resp, _ := c.Post("", &Payload{})
		if !requal(resp, responses[i]) {
			t.Errorf("Post: got\n\t%+v\nexpected\n\t%+v", resp, responses[i])
		}
	}
}

func requal(r1 *Response, r2 *serverResponse) bool {
	return r1.StatusCode == r2.statusCode && r1.Message == r2.body
}
