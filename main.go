package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/mail"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/goji/httpauth"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func sanitizeTarget(target string) string {
	targetRe := regexp.MustCompile("[^A-Za-z0-9_-]")
	return targetRe.ReplaceAllString(target, "")
}
func sanitizeSequence(sequence string) string {
	sequenceRe := regexp.MustCompile("[^A-Z]")
	return sequenceRe.ReplaceAllString(strings.ToUpper(sequence), "")
}

func isIn(num string, params []string) int {
	for i, param := range params {
		if num == param {
			return i
		}
	}

	return -1
}

type Job struct {
	Server      string `json:"server"`
	Target      string `json:"target"`
	Sequence    string `json:"sequence"`
	Email       string `json:"email"`
	ResponseURL string `json:"response"`
}

func ParseConfigName(args []string) (string, []string) {
	resArgs := make([]string, 0)
	file := ""
	for i := 0; i < len(args); i++ {
		if args[i] == "-config" {
			file = args[i+1]
			i++
			continue
		}

		resArgs = append(resArgs, args[i])
	}

	return file, resArgs
}

func main() {
	configFile, args := ParseConfigName(os.Args[1:])

	var config ConfigRoot
	var err error
	if len(configFile) > 0 {
		config, err = ReadConfigFromFile(configFile)

	} else {
		config, err = DefaultConfig()
	}
	if err != nil {
		panic(err)
	}

	err = config.ReadParameters(args)
	if err != nil {
		panic(err)
	}

	log.Println("Using " + config.Mail.Mailer.Type + " mail transport")
	mailer := config.Mail.Mailer.GetTransport()

	r := mux.NewRouter()
	r.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		// var err error
		// if strings.HasPrefix(req.Header.Get("Content-Type"), "multipart/form-data") {
		// 	err = req.ParseMultipartForm(int64(128 * 1024 * 1024))
		// } else {
		// 	err = req.ParseForm()
		// }
		// if err != nil {
		// 	http.Error(w, err.Error(), http.StatusBadRequest)
		// 	return
		// }
		email := req.FormValue("REPLY-E-MAIL")
		server := req.FormValue("SERVER")
		target := sanitizeTarget(req.FormValue("TARGET"))
		sequence := sanitizeSequence(req.FormValue("SEQUENCE"))
		if email == "" || server == "" || target == "" || sequence == "" {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		_, err = mail.ParseAddress(email)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if isIn(server, config.Cameo.Servers) == -1 {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = os.MkdirAll(path.Join(config.Cameo.JobPath, "jobs", server), 0755)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		file, err := os.Create(path.Join(config.Cameo.JobPath, "jobs", server, target+".json"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		err = json.NewEncoder(file).Encode(Job{server, target, sequence, email, config.Cameo.ResponseURL})
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		err = file.Close()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}).Methods("POST")

	r.HandleFunc("/error", func(w http.ResponseWriter, req *http.Request) {
		target := sanitizeTarget(req.FormValue("TARGET"))
		if target == "" {
			http.Error(w, "Invalid target", http.StatusBadRequest)
			return
		}

		err = mailer.Send(Mail{
			config.Mail.Sender,
			"",
			"Error in Target: " + target,
			"Error",
			[]string{},
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}).Methods("POST")

	r.HandleFunc("/success", func(w http.ResponseWriter, req *http.Request) {
		err := req.ParseMultipartForm(int64(128 * 1024 * 1024))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		email := req.FormValue("REPLY-E-MAIL")
		server := req.FormValue("SERVER")
		target := sanitizeTarget(req.FormValue("TARGET"))
		if email == "" || server == "" || target == "" {
			http.Error(w, "Missing parameters", http.StatusBadRequest)
			return
		}

		_, err = mail.ParseAddress(email)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if isIn(server, config.Cameo.Servers) == -1 {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		file, _, err := req.FormFile("FILE")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer file.Close()
		result, err := ioutil.ReadAll(file)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = mailer.Send(Mail{
			config.Mail.Sender,
			email,
			target + " - " + server,
			string(result),
			[]string{},
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}).Methods("POST")

	r.HandleFunc("/jobs", func(w http.ResponseWriter, req *http.Request) {
		server := req.FormValue("SERVER")
		if isIn(server, config.Cameo.Servers) == -1 {
			http.Error(w, "Invalid server", http.StatusBadRequest)
			return
		}
		err := os.MkdirAll(path.Join(config.Cameo.JobPath, "done", server), 0755)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")

		handle, err := os.Open(path.Join(config.Cameo.JobPath, "jobs", server))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		files, err := handle.Readdir(-1)
		handle.Close()
		if err != nil {
			log.Fatal(err)
		}

		for _, file := range files {
			if file.Mode().IsRegular() == false {
				continue
			}
			if filepath.Ext(file.Name()) != "json" {
				continue
			}

			data, err := ioutil.ReadFile(path.Join(config.Cameo.JobPath, "jobs", server, file.Name()))
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.Write([]byte(data))
			w.Write([]byte("\n"))
			err = os.Rename(path.Join(config.Cameo.JobPath, "jobs", server, file.Name()), path.Join("done", server, file.Name()))
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

		}
	}).Methods("POST")

	h := http.Handler(r)
	if config.Server.Auth != nil {
		h = httpauth.SimpleBasicAuth(config.Server.Auth.Username, config.Server.Auth.Password)(h)
	}
	if config.Verbose {
		h = handlers.LoggingHandler(os.Stdout, h)
	}

	srv := &http.Server{
		Handler: h,
		Addr:    config.Server.Address,
	}

	log.Println("CAMEO Responder")
	log.Fatal(srv.ListenAndServe())
}
