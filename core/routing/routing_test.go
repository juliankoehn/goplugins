package routing

import "net/http/httptest"

type (
	user struct {
		ID   int    `json:"id" xml:"id" form:"id" query:"id" param:"id"`
		Name string `json:"name" xml:"name" form:"name" query:"name" param:"name"`
	}
)

const (
	userJSON                    = `{"id":1,"name":"Jon Snow"}`
	userXML                     = `<user><id>1</id><name>Jon Snow</name></user>`
	userForm                    = `id=1&name=Jon Snow`
	invalidContent              = "invalid content"
	userJSONInvalidType         = `{"id":"1","name":"Jon Snow"}`
	userXMLConvertNumberError   = `<user><id>Number one</id><name>Jon Snow</name></user>`
	userXMLUnsupportedTypeError = `<user><>Number one</><name>Jon Snow</name></user>`

	userJSONPretty = `{
  "id": 1,
  "name": "Jon Snow"
}`

	userXMLPretty = `<user>
  <id>1</id>
  <name>Jon Snow</name>
</user>`
)

func request(method, path string, m *Mux) (int, string) {
	req := httptest.NewRequest(method, path, nil)
	rec := httptest.NewRecorder()
	m.ServeHTTP(rec, req)
	return rec.Code, rec.Body.String()
}
