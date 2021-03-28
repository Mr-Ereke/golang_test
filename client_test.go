package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

type TestCase struct {
	req SearchRequest
	err string
}

type Rows struct {
	XMLName xml.Name `xml:"root"`
	Row     []Row    `xml:"row"`
}

type Row struct {
	XMLName   xml.Name `xml:"row"`
	Id        int      `xml:"id"`
	Age       int      `xml:"age"`
	FirstName string   `xml:"first_name"`
	LastName  string   `xml:"last_name"`
	About     string   `xml:"about"`
	Gender    string   `xml:"gender"`
}

func SearchServer(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("AccessToken")

	if token != "secret" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if token != "secret" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	query := r.FormValue("query")
	offset, err := strconv.Atoi(r.FormValue("offset"))

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	limit, err := strconv.Atoi(r.FormValue("limit"))

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	users := ParseXML()

	var filterUsers []User

	if query != "" {
		for _, user := range users {
			if strings.Contains(user.Name, query) || strings.Contains(user.About, query) {
				filterUsers = append(filterUsers, user)
			}
		}
	}

	orderBy, err := strconv.Atoi(r.FormValue("order_by"))

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if orderBy != OrderByAsIs {
		switch orderField := r.FormValue("order_field"); orderField {
		case "Id":
			sort.Slice(filterUsers[:], func(i, j int) bool {
				return filterUsers[i].Id < filterUsers[j].Id
			})
			break
		case "Age":
			sort.Slice(filterUsers[:], func(i, j int) bool {
				return filterUsers[i].Age < filterUsers[j].Age
			})
			break
		case "Name":
			sort.Slice(filterUsers[:], func(i, j int) bool {
				return filterUsers[i].Name < filterUsers[j].Name
			})
			break
		case "":
			sort.Slice(filterUsers[:], func(i, j int) bool {
				return filterUsers[i].Name < filterUsers[j].Name
			})
			break
		default:
			fmt.Errorf("ErrorBadOrderField")
			return
		}
	}

	if orderBy == OrderByAsc {
		for i, j := 0, len(filterUsers)-1; i < j; i, j = i+1, j-1 {
			filterUsers[i], filterUsers[j] = filterUsers[j], filterUsers[i]
		}
	}

	sliceLimit := 0

	if limit <= len(filterUsers) {
		sliceLimit = limit
	} else {
		sliceLimit = len(filterUsers)
	}

	limitedUsers := filterUsers[offset:sliceLimit]

	var response []string

	for _, user := range limitedUsers {
		if jsn, err := json.Marshal(user); err == nil {
			response = append(response, string(jsn))
		}
	}

	io.WriteString(w, "["+strings.Join(response, ",")+"]")
}

func ParseXML() []User {
	xmlFile, err := os.Open("dataset.xml")

	if err != nil {
		fmt.Println(err)
	}

	defer xmlFile.Close()

	byteValue, _ := ioutil.ReadAll(xmlFile)

	var rows Rows
	var users []User

	xml.Unmarshal(byteValue, &rows)

	for i := 0; i < len(rows.Row); i++ {
		newUser := User{}
		newUser.Id = rows.Row[i].Id
		newUser.Name = rows.Row[i].FirstName + rows.Row[i].LastName
		newUser.Age = rows.Row[i].Age
		newUser.About = rows.Row[i].About
		newUser.Gender = rows.Row[i].Gender
		users = append(users, newUser)
	}

	return users
}

func TestRequest(t *testing.T) {
	cases := []TestCase{
		TestCase{
			req: SearchRequest{
				Limit:      26,
				Offset:     0,
				Query:      "minim",
				OrderField: "Id",
				OrderBy:    1,
			},
			err: "",
		},
		TestCase{
			req: SearchRequest{
				Limit:      -1,
				Offset:     0,
				Query:      "minim",
				OrderField: "Id",
				OrderBy:    1,
			},
			err: "",
		},
		TestCase{
			req: SearchRequest{
				Limit:      1,
				Offset:     -1,
				Query:      "minim",
				OrderField: "Id",
				OrderBy:    1,
			},
			err: "",
		},
		TestCase{
			req: SearchRequest{
				Limit:      0,
				Offset:     0,
				Query:      "Nulla cillum enim voluptate consequat laborum esse excepteur occaecat commodo nostrud excepteur ut cupidatat. Occaecat minim incididunt ut proident ad sint nostrud ad laborum sint pariatur. Ut nulla commodo dolore officia. Consequat anim eiusmod amet commodo eiusmod deserunt culpa. Ea sit dolore nostrud cillum proident nisi mollit est Lorem pariatur. Lorem aute officia deserunt dolor nisi aliqua consequat nulla nostrud ipsum irure id deserunt dolore. Minim reprehenderit nulla exercitation labore ipsum",
				OrderField: "Id",
				OrderBy:    1,
			},
			err: "",
		},
		TestCase{
			req: SearchRequest{
				Limit:      5,
				Offset:     0,
				Query:      "Nulla cillum enim voluptate consequat laborum esse excepteur occaecat commodo nostrud excepteur ut cupidatat. Occaecat minim incididunt ut proident ad sint nostrud ad laborum sint pariatur. Ut nulla commodo dolore officia. Consequat anim eiusmod amet commodo eiusmod deserunt culpa. Ea sit dolore nostrud cillum proident nisi mollit est Lorem pariatur. Lorem aute officia deserunt dolor nisi aliqua consequat nulla nostrud ipsum irure id deserunt dolore. Minim reprehenderit nulla exercitation labore ipsum",
				OrderField: "Id",
				OrderBy:    1,
			},
			err: "",
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()

	for _, item := range cases {
		client := &SearchClient{
			AccessToken: "secret",
			URL:         ts.URL,
		}

		_, err := client.FindUsers(item.req)

		if err != nil {
			t.Errorf("[%d] unexpected error: %#v", caseNum, err)
		}
	}
}

func TestUnauthorized(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()

	client := &SearchClient{
		AccessToken: "hack",
		URL:         ts.URL,
	}

	_, err := client.FindUsers(SearchRequest{})

	if err != nil {
	}
}

func TestInternalServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	client := &SearchClient{
		AccessToken: "secret",
		URL:         ts.URL,
	}

	_, err := client.FindUsers(SearchRequest{})

	if err != nil {
	}
}

func TestUnpackJson(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, `{"format":"json"}`)
		return
	}))

	defer ts.Close()

	client := &SearchClient{
		AccessToken: "secret",
		URL:         ts.URL,
	}

	_, err := client.FindUsers(SearchRequest{
		Limit:      3,
		Offset:     0,
		Query:      "minim",
		OrderField: "Id",
		OrderBy:    1,
	})

	if err != nil {
	}
}

func TestBadRequest(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		query := r.FormValue("query")
		io.WriteString(w, query)
	}))

	defer ts.Close()

	client := &SearchClient{
		AccessToken: "secret",
		URL:         ts.URL,
	}

	_, err := client.FindUsers(SearchRequest{
		Limit:      3,
		Offset:     0,
		Query:      "minim",
		OrderField: "Id",
		OrderBy:    1,
	})

	if err != nil {
	}

	_, err = client.FindUsers(SearchRequest{
		Limit:      6,
		Offset:     0,
		Query:      `{"Error":"ErrorBadOrderField"}`,
		OrderField: "Id",
		OrderBy:    1,
	})

	if err != nil {
	}

	_, err = client.FindUsers(SearchRequest{
		Limit:      2,
		Offset:     0,
		Query:      `{"Error":"unknown bad request error"}`,
		OrderField: "Id",
		OrderBy:    1,
	})

	if err != nil {
	}
}

func TestTimeoutError(t *testing.T) {
	responseHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2000 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(""))
	})
	ts := httptest.NewServer(responseHandler)
	defer ts.Close()

	client := &SearchClient{
		AccessToken: "secret",
		URL:         ts.URL,
	}

	_, err := client.FindUsers(SearchRequest{
		Limit:      3,
		Offset:     0,
		Query:      "minim",
		OrderField: "Id",
		OrderBy:    1,
	})
	if err != nil {
	}
}

func TestServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	}))

	defer ts.Close()

	client := &SearchClient{
		AccessToken: "secret",
		URL:         "error",
	}

	_, err := client.FindUsers(SearchRequest{
		Limit:      3,
		Offset:     0,
		Query:      "minim",
		OrderField: "Id",
		OrderBy:    1,
	})

	if err != nil {
		fmt.Println(err)
	}
}
