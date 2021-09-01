package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
)

var defaultFileContent = []byte(`{
    // should mmseqs und webserver output be printed
    "verbose": true,
    "server" : {
        "address"    : "127.0.0.1:8081",
        // prefix for all API endpoints
        "pathprefix" : "/api/",
        /* enable HTTP Basic Auth (optional)
        "auth": {
            "username" : "",
            "password" : ""
        },
        */
        // should CORS headers be set to allow requests from anywhere
        "cors"       : true
    },
    // camep specific settigns, special character ~ is resolved relative to the binary location
    "cameo" : {
        // path to mmseqs databases, has to be shared between server/workers
        "path"    : "~",
        // server names
        "servers" : [],
		// response URL
		"response" : ""
    },
    "mail" : {
        "mailer" : {
            // three types available:
            // null: uses NullTransport class, which ignores all sent emails
            "type" : "null"
            /* smtp: Uses SMTP to send emails example for gmail
            "type" : "smtp",
            "transport" : {
                // full host URL with port
                "host" : "smtp.gmail.com:587",
                // RFC 4616  PLAIN authentication
                "auth" : {
                    {
                        // empty for gmail
                        "identity" : "",
                        // gmail user
                        "username" : "user@gmail.com",
                        "password" : "password",
                        "host" : "smtp.gmail.com",
                    }
                }
            }
            */
            /* mailgun: Uses the mailgun API to send emails
            "type"      : "mailgun",
            "transport" : {
                // mailgun domain
                "domain" : "mail.mmseqs.com",
                // mailgun API private key
                "secretkey" : "key-XXXX",
                // mailgun API public key
                "publickey" : "pubkey-XXXX"
            }
            */
        },
        // Email FROM field
        "sender"    : "mail@example.org",
        /* Bracket notation is also possible:
        "sender"    : "Webserver <mail@example.org>",
        */
    }
}
`)

type ConfigCameo struct {
	JobPath     string   `json:"path"`
	Servers     []string `json:"servers"`
	ResponseURL string   `json:"response"`
}

type ConfigMail struct {
	Mailer *ConfigMailtransport `json:"mailer"`
	Sender string               `json:"sender"`
	Target string               `json:"target"`
}

type ConfigAuth struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type ConfigServer struct {
	Address    string      `json:"address" validate:"required"`
	PathPrefix string      `json:"pathprefix"`
	CORS       bool        `json:"cors"`
	Auth       *ConfigAuth `json:"auth"`
}

type ConfigRoot struct {
	Server  ConfigServer `json:"server" validate:"required"`
	Cameo   ConfigCameo  `json:"cameo" validate:"required"`
	Mail    ConfigMail   `json:"mail"`
	Verbose bool         `json:"verbose"`
}

func ReadConfigFromFile(name string) (ConfigRoot, error) {
	file, err := os.Open(name)
	if err != nil {
		return ConfigRoot{}, err
	}
	defer file.Close()

	absPath, err := filepath.Abs(name)
	if err != nil {
		return ConfigRoot{}, err
	}

	relativeTo := filepath.Dir(absPath)

	return ReadConfig(file, relativeTo)
}

func DefaultConfig() (ConfigRoot, error) {
	r := bytes.NewReader(defaultFileContent)

	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	relativeTo := filepath.Dir(ex)

	return ReadConfig(r, relativeTo)
}

func ReadConfig(r io.Reader, relativeTo string) (ConfigRoot, error) {
	var config ConfigRoot
	if err := DecodeJsonAndValidate(r, &config); err != nil {
		return config, fmt.Errorf("Fatal error for config file: %s\n", err)
	}

	paths := []*string{&config.Cameo.JobPath}
	for _, path := range paths {
		if strings.HasPrefix(*path, "~") {
			*path = strings.TrimLeft(*path, "~")
			*path = filepath.Join(relativeTo, *path)
		}
	}

	return config, nil
}

func (c *ConfigRoot) CheckPaths() error {
	paths := []string{c.Cameo.JobPath}
	for _, path := range paths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			os.MkdirAll(path, 0755)
		}
	}
	return nil
}

func (c *ConfigRoot) ReadParameters(args []string) error {
	var key string
	inParameter := false
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			if inParameter == true {
				return errors.New("Invalid Parameter String")
			}
			key = strings.TrimLeft(arg, "-")
			inParameter = true
		} else {
			if inParameter == false {
				return errors.New("Invalid Parameter String")
			}
			err := c.setParameter(key, arg)
			if err != nil {
				return err
			}
			inParameter = false
		}
	}

	if inParameter == true {
		return errors.New("Invalid Parameter String")
	}

	return nil
}

func (c *ConfigRoot) setParameter(key string, value string) error {
	path := strings.Split(key, ".")
	return setNodeValue(c, path, value)
}

// DFS in Config Tree to set the new value
func setNodeValue(node interface{}, path []string, value string) error {
	if len(path) == 0 {
		if v, ok := node.(reflect.Value); ok {
			if v.IsValid() == false || v.CanSet() == false {
				return errors.New("Leaf node is not valid")
			}

			switch v.Kind() {
			case reflect.Struct:
				return errors.New("Leaf node is a struct")
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				i, err := strconv.ParseInt(value, 10, 64)
				if err != nil {
					return err
				}
				v.SetInt(i)
				break
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				i, err := strconv.ParseUint(value, 10, 64)
				if err != nil {
					return err
				}
				v.SetUint(i)
				break
			case reflect.Bool:
				b, err := strconv.ParseBool(value)
				if err != nil {
					return err
				}
				v.SetBool(b)
				break
			case reflect.String:
				v.SetString(value)
				break
			default:
				return errors.New("Leaf node type not implemented")
			}
			return nil
		} else {
			return errors.New("Leaf node is not a value")
		}
	}

	v, ok := node.(reflect.Value)
	if !ok {
		v = reflect.ValueOf(node).Elem()
	}

	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			t := v.Type().Elem()
			n := reflect.New(t)
			v.Set(n)
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return errors.New("Node is not a struct")
	}

	for i := 0; i < v.NumField(); i++ {
		tag := v.Type().Field(i).Tag.Get("json")
		if tag == "" || tag == "-" {
			continue
		}

		if tag == path[0] {
			f := v.Field(i)
			return setNodeValue(f, path[1:], value)
		}
	}

	return errors.New("Path not found in config")
}
