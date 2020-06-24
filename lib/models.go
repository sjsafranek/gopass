package lib

import (
	"encoding/json"
	"fmt"
	"io"
	// "time"
)

const VERSION = "0.0.1"
const DEFAULT_NAMESPACE = "store"

// Request
type Request struct {
	Id      string        `json:"id,ompitempty"`
	Version string        `json:"version"`
	Method  string        `json:"method,omitempty"`
	Params  RequestParams `json:"params,omitempty"`
}

type RequestParams struct {
	Namespace  string `json:"namespace"`
	Key        string `json:"key"`
	Value      string `json:"value"`
	Passphrase string `json:"passphrase"`
}

func (self *Request) Unmarshal(data string) error {
	return json.Unmarshal([]byte(data), self)
}

// Response
type Response struct {
	Id      string       `json:"id,omitempty"`
	Version string       `json:"version,omitempty"`
	Status  string       `json:"status"`
	Error   string       `json:"error,omitempty"`
	Data    *ResponseData `json:"data,omitempty"`
}

func (self *Response) SetError(err error) {
	self.Status = "error"
	self.Error = err.Error()
}

func (self *Response) SetSuccess(err error) {
	self.Status = "ok"
	self.Error = ""
}

func (self *Response) Marshal() (string, error) {
	b, err := json.Marshal(self)
	if nil != err {
		return "", err
	}
	return string(b), err
}

func (self *Response) Write(w io.Writer) error {
	payload, err := self.Marshal()
	if nil != err {
		return err
	}
	_, err = fmt.Fprintln(w, payload)
	return err
}

func (self *Response) Print() error {
	payload, err := self.Marshal()
	if nil != err {
		return err
	}
	fmt.Println(payload)
	return nil
}

type ResponseData struct {
	Key        string    `json:"key,omitempty"`
	Keys       *[]string `json:"keys,omitempty"`
	Value      string    `json:"value,omitempty"`
	Namespace  string    `json:"namespace,omitempty"`
	Namespaces *[]string `json:"namespaces,omitempty"`
	Passphrase string    `json:"passphrase,omitempty"`
}

func NewResponse() *Response {
	// return &Response{Version: VERSION, Id: NewId(), Status: "ok"}
	return &Response{Version: VERSION, Status: "ok"}
}
