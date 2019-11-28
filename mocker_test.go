package mocker

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/ricardogama/api-mocker-go/mocks"
	. "github.com/smartystreets/goconvey/convey"
)

func Test_New(t *testing.T) {
	Convey("New()", t, func() {
		Convey("Returns a mocker instance with given arguments", func() {
			mocker := New("foo")

			So(mocker, ShouldResemble, &Mocker{
				BasePath: "foo",
			})
		})
	})
}

func Test_Mocker_Results(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	handlerMock := mocks.NewMockHandler(mockCtrl)

	Convey("*Mocker.Results()", t, func() {
		Convey("Returns an error when it's not possible to create a request", func() {
			mock := &Mocker{BasePath: ":"}
			results, err := mock.Results()

			So(err.Error(), ShouldEqual, "parse :/mocks: missing protocol scheme")
			So(results, ShouldBeNil)
		})

		Convey("Returns an error when it's not possible to execute the request", func() {
			mock := &Mocker{BasePath: "foo://bar"}
			results, err := mock.Results()

			So(err.Error(), ShouldEqual, `Get foo://bar/mocks: unsupported protocol scheme "foo"`)
			So(results, ShouldBeNil)
		})

		Convey("Returns an error when the request does not return 200", func() {
			server := httptest.NewServer(handlerMock)
			defer server.Close()

			var req *http.Request
			handlerMock.EXPECT().
				ServeHTTP(gomock.Any(), gomock.Any()).
				Do(func(w http.ResponseWriter, r *http.Request) {
					req = r
					w.WriteHeader(500)
				}).
				Times(1)

			mock := &Mocker{BasePath: server.URL}
			results, err := mock.Results()

			So(err.Error(), ShouldEqual, "failed to get mocks")
			So(results, ShouldBeNil)
			So(req.Method, ShouldEqual, "GET")
			So(req.URL.Path, ShouldResemble, "/mocks")
		})

		Convey("Returns an error when the response cannot be decoded", func() {
			server := httptest.NewServer(handlerMock)
			defer server.Close()

			var req *http.Request
			handlerMock.EXPECT().
				ServeHTTP(gomock.Any(), gomock.Any()).
				Do(func(w http.ResponseWriter, r *http.Request) {
					req = r
					fmt.Fprintf(w, "foo")
				}).
				Times(1)

			mock := &Mocker{BasePath: server.URL}
			results, err := mock.Results()

			So(err.Error(), ShouldEqual, "invalid character 'o' in literal false (expecting 'a')")
			So(results, ShouldBeNil)
			So(req.Method, ShouldEqual, "GET")
			So(req.URL.Path, ShouldResemble, "/mocks")
		})

		Convey("Returns the results", func() {
			server := httptest.NewServer(handlerMock)
			defer server.Close()

			var req *http.Request
			handlerMock.EXPECT().
				ServeHTTP(gomock.Any(), gomock.Any()).
				Do(func(w http.ResponseWriter, r *http.Request) {
					req = r
					fmt.Fprintf(w, `{"expected":[],"unexpected":[]}`)
				})

			mock := &Mocker{BasePath: server.URL}
			results, err := mock.Results()

			So(err, ShouldBeNil)
			So(req.Method, ShouldEqual, "GET")
			So(req.URL.Path, ShouldResemble, "/mocks")
			So(results, ShouldResemble, &Results{
				Expected:   []*Request{},
				Unexpected: []*Request{},
			})
		})
	})
}

func Test_Mocker_Ensure(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	handlerMock := mocks.NewMockHandler(mockCtrl)

	Convey("*Mocker.Ensure()", t, func() {
		Convey("Returns an error when there are expected requests", func() {
			server := httptest.NewServer(handlerMock)
			defer server.Close()

			var req *http.Request
			handlerMock.EXPECT().
				ServeHTTP(gomock.Any(), gomock.Any()).
				Do(func(w http.ResponseWriter, r *http.Request) {
					req = r
					fmt.Fprintf(
						w,
						`
							{
								"expected": [{}]
							}
						`,
					)
				}).
				Times(1)

			mock := &Mocker{BasePath: server.URL}
			err := mock.Ensure()

			So(strings.Join(strings.Fields(err.Error()), " "), ShouldEqual, `missing 1 expected calls: [ { "method": "", "path": "", "response": null } ]`)
			So(req.Method, ShouldEqual, "GET")
			So(req.URL.Path, ShouldResemble, "/mocks")
		})

		Convey("Returns an error when there are unexpected requests", func() {
			server := httptest.NewServer(handlerMock)
			defer server.Close()

			var req *http.Request
			handlerMock.EXPECT().
				ServeHTTP(gomock.Any(), gomock.Any()).
				Do(func(w http.ResponseWriter, r *http.Request) {
					req = r
					fmt.Fprintf(
						w,
						`
							{
								"unexpected": [{}]
							}
						`,
					)
				}).
				Times(1)

			mock := &Mocker{BasePath: server.URL}
			err := mock.Ensure()

			So(strings.Join(strings.Fields(err.Error()), " "), ShouldEqual, `1 unexpected calls: [ { "method": "", "path": "", "response": null } ]`)
			So(req.Method, ShouldEqual, "GET")
			So(req.URL.Path, ShouldResemble, "/mocks")
		})

		Convey("Does not fail when there are no expected and unexpected requests", func() {
			server := httptest.NewServer(handlerMock)
			defer server.Close()

			var req *http.Request
			handlerMock.EXPECT().
				ServeHTTP(gomock.Any(), gomock.Any()).
				Do(func(w http.ResponseWriter, r *http.Request) {
					req = r
					fmt.Fprintf(
						w,
						`
							{
								"unexpected": [],
								"expected": []
							}
						`,
					)
				}).
				Times(1)

			mock := &Mocker{BasePath: server.URL}
			err := mock.Ensure()

			So(err, ShouldBeNil)
			So(req.Method, ShouldEqual, "GET")
			So(req.URL.Path, ShouldResemble, "/mocks")
		})
	})
}

func Test_Mocker_Expect(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	handlerMock := mocks.NewMockHandler(mockCtrl)

	Convey("*Mocker.Expect()", t, func() {
		Convey("Returns an error when it's not possible to create a request", func() {
			mock := &Mocker{BasePath: ":"}
			err := mock.Expect(nil)

			So(err.Error(), ShouldEqual, "parse :/mocks: missing protocol scheme")
		})

		Convey("Returns an error when it's not possible to execute the request", func() {
			mock := &Mocker{}
			err := mock.Expect(nil)

			So(err.Error(), ShouldEqual, `Post /mocks: unsupported protocol scheme ""`)
		})

		Convey("Does not fail when the request returns 201", func() {
			server := httptest.NewServer(handlerMock)
			defer server.Close()

			var req *http.Request
			var body []byte
			handlerMock.EXPECT().
				ServeHTTP(gomock.Any(), gomock.Any()).
				Do(func(w http.ResponseWriter, r *http.Request) {
					req = r
					body, _ = ioutil.ReadAll(req.Body)
					w.WriteHeader(201)
				}).
				Times(1)

			mock := &Mocker{BasePath: server.URL}
			err := mock.Expect(&Request{
				Query: map[string]string{
					"biz": "baz",
				},
				Method: "foo",
				Path:   "bar",
				Response: &Response{
					Status: 200,
				},
			})

			So(err, ShouldBeNil)
			So(req.Method, ShouldEqual, "POST")
			So(req.URL.Path, ShouldResemble, "/mocks")
			So(req.Header["Content-Type"], ShouldResemble, []string{"application/json"})
			So(string(body), ShouldEqual, `{"method":"foo","path":"bar","query":{"biz":"baz"},"response":{"status":200}}`)
		})

		Convey("It fails when the request does not return 201", func() {
			server := httptest.NewServer(handlerMock)
			defer server.Close()

			var req *http.Request
			var body []byte
			handlerMock.EXPECT().
				ServeHTTP(gomock.Any(), gomock.Any()).
				Do(func(w http.ResponseWriter, r *http.Request) {
					req = r
					body, _ = ioutil.ReadAll(req.Body)
					w.WriteHeader(500)
					fmt.Fprintf(w, `{"err":"damn"}`)
				}).
				Times(1)

			mock := &Mocker{BasePath: server.URL}
			err := mock.Expect(&Request{
				Query: map[string]string{
					"biz": "baz",
				},
				Method: "foo",
				Path:   "bar",
				Response: &Response{
					Status: 200,
				},
			})

			So(strings.Join(strings.Fields(err.Error()), " "), ShouldEqual, `failed to create mock { "err": "damn" }`)
			So(req.Method, ShouldEqual, "POST")
			So(req.URL.Path, ShouldResemble, "/mocks")
			So(req.Header["Content-Type"], ShouldResemble, []string{"application/json"})
			So(string(body), ShouldEqual, `{"method":"foo","path":"bar","query":{"biz":"baz"},"response":{"status":200}}`)
		})
	})
}

func Test_Mocker_Clear(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	handlerMock := mocks.NewMockHandler(mockCtrl)

	Convey("*Mocker.Clear()", t, func() {
		Convey("Returns an error when it's not possible to create a request", func() {
			mock := &Mocker{BasePath: ":"}
			err := mock.Clear()

			So(err.Error(), ShouldEqual, "parse :/mocks: missing protocol scheme")
		})

		Convey("Returns an error when it's not possible to execute the request", func() {
			mock := &Mocker{}
			err := mock.Clear()

			So(err.Error(), ShouldEqual, `Delete /mocks: unsupported protocol scheme ""`)
		})

		Convey("It fails when the request does not return 200", func() {
			server := httptest.NewServer(handlerMock)
			defer server.Close()

			var req *http.Request
			handlerMock.EXPECT().
				ServeHTTP(gomock.Any(), gomock.Any()).
				Do(func(w http.ResponseWriter, r *http.Request) {
					req = r
					w.WriteHeader(500)
				}).
				Times(1)

			mock := &Mocker{BasePath: server.URL}
			err := mock.Clear()

			So(err.Error(), ShouldEqual, "failed to clear mocks")
			So(req.Method, ShouldEqual, "DELETE")
			So(req.URL.Path, ShouldResemble, "/mocks")
		})

		Convey("Does not fail when the request returns 204", func() {
			server := httptest.NewServer(handlerMock)
			defer server.Close()

			var req *http.Request
			handlerMock.EXPECT().
				ServeHTTP(gomock.Any(), gomock.Any()).
				Do(func(w http.ResponseWriter, r *http.Request) {
					req = r
					w.WriteHeader(204)
				}).
				Times(1)

			mock := &Mocker{BasePath: server.URL}
			err := mock.Clear()

			So(err, ShouldBeNil)
			So(req.Method, ShouldEqual, "DELETE")
			So(req.URL.Path, ShouldResemble, "/mocks")
		})
	})
}
