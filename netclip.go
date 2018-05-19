package main

import (
  "io"
  "os"
  "fmt"
  "flag"
  "bufio"
  "net/http"
  "net/url"
  "strings"
  "io/ioutil"
  "encoding/json"
  "github.com/op/go-logging"
  "github.com/mitchellh/go-homedir"
)

var app_version = "0.1.0"
var key_file, err = homedir.Expand("~/.netclip_key")
var base_url = "http://localhost:8000/"

var log = logging.MustGetLogger("netclip")

var (
        key string
        register bool
        version bool
)

type ServerResponse struct {
        ResponseCode string
        Message string
}

func postData(endpoint string, data url.Values) ServerResponse {
        resp, err := http.PostForm(base_url + endpoint, data)
        if err != nil { log.Error("%v", err) }

        defer resp.Body.Close()
        body, err:= ioutil.ReadAll(resp.Body)

        if err != nil { log.Error("%v", err) }

        var response ServerResponse
        err = json.Unmarshal(body, &response)

        if err != nil { log.Error("%v", err) }

        if response.ResponseCode == "200" {
                log.Info("200 OK")
        } else {
                log.Error("%s: %s", response.ResponseCode, response.Message)
        }

        return response
}

func getData(endpoint string) ServerResponse {
	resp, err := http.Get(base_url + endpoint)
	if err != nil { log.Error("%v", err) }

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
        if err != nil { log.Error("%v", err) }

        var response ServerResponse
        err = json.Unmarshal(body, &response)

        if err != nil { log.Error("%v", err) }

        if response.ResponseCode == "200" {
                log.Info("200 OK")
        } else {
                log.Error("%s: %s", response.ResponseCode, response.Message)
        }

        return response
}

func main() {
        flag.StringVar(&key, "key", "", "Set the API Key")
        flag.BoolVar(&register, "register", false, "Generate a new API key")
        flag.BoolVar(&version, "version", false, "Show Version")

        flag.Parse()

        if version {
                fmt.Println(app_version)
                os.Exit(0)
        }

        if register {
                v := url.Values{}
                v.Add("action", "register")
                res := postData("register", v)
                fmt.Println(res.Message)
                os.Exit(0)
        }

        if key != "" {
                // The key file is read only, so we will remove it before attempting to write it out again.
                os.Remove(key_file)

                // Now we will write the provided key back out to the file
                err := ioutil.WriteFile(key_file, []byte(key), 0400)
                if err != nil { fmt.Printf("Error: %v\n", err) }

        } else if _, err := os.Stat(key_file); err == nil {
                b, err := ioutil.ReadFile(key_file)
                if err != nil { log.Error("%v", err) }
                key = string(b)

        } else {
                fmt.Println("No API key set. Set with netclip -key [key]. If you don't have a key, first call netclip -register to obtain one.")
                os.Exit(1)
        }

        fi, err := os.Stdin.Stat()
        if err != nil {
                panic(err)
        }

	log_backend, err := logging.NewSyslogBackend("netclip")

	if err != nil { fmt.Printf("Error: %", err) }
	logging.SetBackend(log_backend)

        var format = logging.MustStringFormatter("%{level} %{message}")
        logging.SetFormatter(format)
        logging.SetLevel(logging.INFO, "netclip")

        if fi.Mode() & os.ModeNamedPipe == 0 {
                // Here I'll want to fetch the last clip and spit it out
		res := getData("clip/get/" + key)
                fmt.Println(res.Message)
        } else {
                reader := bufio.NewReader(os.Stdin)

                var clip string

                for {
                        input, err := reader.ReadString('\n')
                        if err != nil && err == io.EOF {
                                break
                        }

                        clip += input
                }

		// Trim off trailing newline
		clip = strings.TrimSpace(clip)

                v := url.Values{}
                v.Add("action", "clip")
                v.Add("clip", clip)

                res := postData("clip/save/" + key, v)
                fmt.Println(res.Message)
        }
}

