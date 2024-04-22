package main

import (
	"crypto/sha256"
	"encoding/base64"
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
    "strconv"

	"github.com/goji/httpauth"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func sanitizeTarget(value string) string {
	rx := regexp.MustCompile("[^A-Za-z0-9_-]")
	return rx.ReplaceAllString(value, "")
}

func sanitizeStoichiometry(value string) string {
	rx := regexp.MustCompile("[^A-Za-z0-9]")
	return rx.ReplaceAllString(value, "")
}

func sanitizeSequence(value string) string {
    rx := regexp.MustCompile("[^A-Z:]")
	return rx.ReplaceAllString(strings.ToUpper(value), "")
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
	Server        string `json:"server"`
	Target        string `json:"target"`
	Sequence      string `json:"sequence"`
    Stoichiometry string `json:"stoichiometry"`
	Email         string `json:"email"`
	ResponseURL   string `json:"response"`
}

func (r Job) Hash() string {
	h := sha256.New224()
	h.Write([]byte(r.Server))
	h.Write([]byte(r.Target))
	h.Write([]byte(r.Sequence))
	h.Write([]byte(r.Stoichiometry))
	h.Write([]byte(r.Email))
	h.Write([]byte(r.ResponseURL))

	bs := h.Sum(nil)
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(bs)
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

    splitStoichiometry := regexp.MustCompile("[0-9]+")
	r := mux.NewRouter()
	r.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		email := req.FormValue("REPLY-E-MAIL")
		server := req.FormValue("SERVER")
		target := sanitizeTarget(req.FormValue("TARGET"))
		sequence := req.FormValue("SEQUENCE")
        stoichiometry := sanitizeStoichiometry(req.FormValue("STOICHIOMETRY"))
        if stoichiometry != "" {
            repeats := splitStoichiometry.FindAllString(stoichiometry, -1)
            i := 0
            var s []string
            for _, line := range strings.Split(strings.TrimSuffix(sequence, "\n"), "\n") {
                if strings.HasPrefix(line, ">") {
                    continue
                }
                repeat, _ := strconv.Atoi(repeats[i])
                for j := 0; j < repeat; j++ {
                    s = append(s, line)
                }
                i += 1
            }
            sequence = strings.Join(s, ":")
        }
        sequence = sanitizeSequence(sequence)
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

		job := Job{server, target, sequence, stoichiometry, email, config.Cameo.ResponseURL}
		file, err := os.Create(path.Join(config.Cameo.JobPath, "jobs", server, target+"."+job.Hash()+".json"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		err = json.NewEncoder(file).Encode(job)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		err = file.Close()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		confirmEmail := email
		if config.Mail.Confirm != "" {
			confirmEmail = config.Mail.Confirm
		}

		err = mailer.Send(Mail{
			config.Mail.Sender,
			confirmEmail,
			target + " - query received by " + server,
			"",
			config.Mail.BCC,
		})
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
			config.Mail.BCC,
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
			config.Mail.BCC,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}).Methods("POST")

	r.HandleFunc("/jobs", func(w http.ResponseWriter, req *http.Request) {
        for _, server := range config.Cameo.Servers {
            err = os.MkdirAll(path.Join(config.Cameo.JobPath, "jobs", server), 0755)
            if err != nil {
                http.Error(w, err.Error(), http.StatusBadRequest)
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
                if filepath.Ext(file.Name()) != ".json" {
                    continue
                }

                data, err := ioutil.ReadFile(path.Join(config.Cameo.JobPath, "jobs", server, file.Name()))
                if err != nil {
                    http.Error(w, err.Error(), http.StatusBadRequest)
                    return
                }
                w.Write([]byte(data))
                if data[len(data)-1] != '\n' {
                    w.Write([]byte("\n"))
                }
                err = os.Rename(path.Join(config.Cameo.JobPath, "jobs", server, file.Name()), path.Join("done", server, file.Name()))
                if err != nil {
                    http.Error(w, err.Error(), http.StatusBadRequest)
                    return
                }
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
