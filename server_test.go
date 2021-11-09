package gorouter

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func dummyHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, p PathParams) {
	w.Write([]byte("ACK"))
}

type contextKey string

func dummyHandlerContext(ctx context.Context, w http.ResponseWriter, r *http.Request, p PathParams) {
	if v := ctx.Value(contextKey("di")); v != nil {
		w.Write([]byte("ACK"))
	}
}

func TestHTTPServer_ValidRoute(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/unique/path/123/value", nil)

	routes := CompileRoutes([]Route{
		{"GET", "/unique/path/:id/value", dummyHandler},
	})

	server := Server{CompiledRoutes: routes}
	server.ServeHTTP(w, r)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Error("HTTPServer did not return an OK status on valid request")
	}
	if string(body) != "ACK" {
		t.Error("HTTPServer did not return the proper body for the request")
	}
}

func TestHTTPServer_MethodNotAllowed(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/myPath", nil)

	routes := CompileRoutes([]Route{
		{"GET", "/myPath", dummyHandler},
	})

	server := Server{CompiledRoutes: routes}
	server.ServeHTTP(w, r)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Error("HTTPServer did not reject method not allowed correctly")
	}
	if string(body) != "Method not allowed\n" {
		t.Error("HTTPServer did not provide error for method not allowed")
	}
}

func TestHTTPServer_PathNotFound(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/wrongPath", nil)

	routes := CompileRoutes([]Route{
		{"GET", "/myPath", dummyHandler},
	})

	server := Server{CompiledRoutes: routes}
	server.ServeHTTP(w, r)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusNotFound {
		t.Error("HTTPServer did not reject path not found correctly")
	}
	if string(body) != "Path not found\n" {
		t.Error("HTTPServer did not provide error for path not found")
	}
}

func TestHTTPServer_MiddlewarePreservesContext(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/myPath", nil)

	routes := CompileRoutes([]Route{
		{"GET", "/myPath", dummyHandlerContext},
	})
	var middlewareFunc MiddlewareFunc = func(w http.ResponseWriter, r *http.Request, p PathParams, h HandlerFunc) {
		ctx := context.WithValue(context.Background(), contextKey("di"), "true")
		h(ctx, w, r, p)
	}
	server := Server{CompiledRoutes: routes, Middleware: middlewareFunc}
	server.ServeHTTP(w, r)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		t.Error("HTTPServer failed to serve middleware")
	}
	if string(body) != "ACK" {
		t.Error("HTTPServer failed to pass context in middlware")
	}
}
