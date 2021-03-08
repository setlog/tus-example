package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/tus/tusd/pkg/filestore"
	tusd "github.com/tus/tusd/pkg/handler"
)

func main() {
	store := filestore.New("./uploads")

	composer := tusd.NewStoreComposer()
	store.UseIn(composer)

	handler, err := tusd.NewHandler(tusd.Config{
		BasePath:              "/files/",
		StoreComposer:         composer,
		NotifyCompleteUploads: true,
		PreUploadCreateCallback: func(hook tusd.HookEvent) error {
			hook.Upload.MetaData["filename"] = hook.HTTPRequest.Header.Get("filename")
			return nil
		},
	})
	if err != nil {
		panic(fmt.Errorf("Unable to create handler: %s", err))
	}

	go func() {
		for {
			event := <-handler.CompleteUploads
			fmt.Printf("Upload %s finished\n", event.Upload.ID)
		}
	}()

	customHandler := http.StripPrefix("/files/", handler)
	handler.Middleware(customHandler)
	http.Handle("/files/", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		err := checkJWT(req.Header.Get("Authorization"))
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
		}
		customHandler.ServeHTTP(w, req)
	}))

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(fmt.Errorf("Unable to listen: %s", err))
	}
}

func checkJWT(authorizationHeader string) error {
	jwt := strings.TrimLeft(authorizationHeader, "Bearer ")
	if jwt != "TrueJWT" {
		return fmt.Errorf("Access denied")
	}
	return nil
}
