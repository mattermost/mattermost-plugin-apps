package function

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

func Handle(w http.ResponseWriter, req *http.Request) {
	var input []byte

	if req.Body != nil {
		defer req.Body.Close()

		body, _ := ioutil.ReadAll(req.Body)

		input = body
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Body: %s", string(input))))
}
