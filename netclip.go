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
	printHelp bool
        key string
        register bool
        version bool
)

type ServerResponse struct {
        ResponseCode string
        Message string
}

func readKey() string {
	if _, err := os.Stat(key_file); err == nil {
		b, err := ioutil.ReadFile(key_file)
		if err != nil { log.Error("%v", err) }
		return string(b)
	}

	return ""
}

func writeKey(new_key string) bool {
	// The key file is read only, so we will remove it before attempting to write it out again.
	os.Remove(key_file)

	// Now we will write the provided key back out to the file
	err := ioutil.WriteFile(key_file, []byte(new_key), 0400)
	if err != nil { 
		fmt.Printf("Error: %v\n", err) 
		return false
	}

	return true
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
	log_backend, err := logging.NewSyslogBackend("netclip")

	if err != nil { fmt.Printf("Error: %", err) }

	logging.SetBackend(log_backend)
        var format = logging.MustStringFormatter("%{level} %{message}")
        logging.SetFormatter(format)
        logging.SetLevel(logging.INFO, "netclip")

	flag.BoolVar(&printHelp, "help", true, "Print this help message.")
        flag.StringVar(&key, "key", "", "Set the API Key")
        flag.BoolVar(&register, "register", false, "Generate a new API key")
        flag.BoolVar(&version, "version", false, "Show Version")

        flag.Parse()

	if printHelp {
		fmt.Println("-----------------------------")
		flag.PrintDefaults()
		fmt.Println("-----------------------------")
		os.Exit(0)
        }

        if version {
                fmt.Println(app_version)
                os.Exit(0)
        }

        if register {
		saved_key := readKey()
		if saved_key != "" {
			fmt.Println("You have already registered this machine. Your API key is: " + saved_key)
		} else {
			v := url.Values{}
			v.Add("action", "register")
			res := postData("register", v)

			writeKey(res.Message)
			fmt.Println("Your API key is: " + res.Message)
		}
                os.Exit(0)
        }

	saved_key := readKey()

        if key != "" {
		writeKey(key)
		fmt.Println("Your API key has been set.")
		os.Exit(0)
        } else if saved_key != "" {
                key = saved_key

        } else {
                fmt.Println("No API key set. Set with netclip -key [key]. If you don't have a key, first call netclip -register to obtain one.")
                os.Exit(1)
        }

        fi, err := os.Stdin.Stat()
        if err != nil {
                panic(err)
        }

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

