package pusher

import (
	"bytes"
	"io/ioutil"
	"net/http"
)

type Request struct {
	method string
	url    string
	body   []byte
}

func (r *Request) send() (error, string) {
	req, err := http.NewRequest(r.method, r.url, bytes.NewBuffer(r.body))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return err, "Error"
	}
	defer resp.Body.Close()
	resp_body, _ := ioutil.ReadAll(resp.Body)
	return nil, string(resp_body)
}
