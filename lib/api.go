package lib

import (
	"encoding/json"
)

func NewApi(db Database) *Api {
	return &Api{db: db}
}

type Api struct {
	db Database
}

func (self *Api) DoJSON(jdata string) (*Response, error) {
	var request Request
	err := json.Unmarshal([]byte(jdata), &request)
	if nil != err {
		response := NewResponse()
		response.SetError(err)
		return response, err
	}
	return self.Do(&request)
}

func (self *Api) Do(request *Request) (*Response, error) {
	response := NewResponse()
	response.Id = request.Id

	if "" == request.Params.Namespace {
		request.Params.Namespace = DEFAULT_NAMESPACE
	}

	err := func() error {
		switch request.Method {

		case "get":
			value, err := self.db.Get(request.Params.Namespace, request.Params.Key, request.Params.Passphrase)
			if nil != err {
				response.SetError(err)
			}
			response.Data = &ResponseData{Value: value}

		case "set":
			err := self.db.Set(request.Params.Namespace, request.Params.Key, request.Params.Value, request.Params.Passphrase)
			if nil != err {
				response.SetError(err)
			}

		case "delete":
			err := self.db.Del(request.Params.Namespace, request.Params.Key, request.Params.Passphrase)
			if nil != err {
				response.SetError(err)
			}

		case "get_keys":
			results, err := self.db.Keys(request.Params.Namespace)
			if nil != err {
				response.SetError(err)
			}
			response.Data = &ResponseData{Keys: &results}

		case "get_namespaces":
			results, err := self.db.Tables()
			if nil != err {
				response.SetError(err)
			}
			response.Data = &ResponseData{Namespaces: &results}

		case "create_namespace":
			err := self.db.CreateTable(request.Params.Namespace)
			if nil != err {
				response.SetError(err)
			}
		}

		return nil
	}()

	return response, err
}
