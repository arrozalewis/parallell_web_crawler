package main

import(
  "os"
  "fmt"
  "log"
  "sync"
  "bytes"
  "bufio"
  "runtime"
  "strings"
  "net/http"
  "io/ioutil"
  "encoding/json"
  "golang.org/x/net/html"
)

type info struct {
  Title     string `json: title`
  Location  string `json: loc`
  Company   string `json: comp`
  Url       string `json: url`
}

type jsReq struct {
  Listings []string `json: listings`
}

type jsResp struct {
  Tlcu  []info `json: tlcu`
}

type Comparator  func(n * html.Node) bool

const input   int    = 1
const output  int    = 2
const chans   int    = 2
const offset  int    = 2
const reqMeth string = "POST"
const reqUrl  string = "http://localhost:8080/get_jobs"
const hKey    string = "Content-Type"
const hVal    string = "application/json"
//1st <div>elem w/ company attr holds company name
const company string = "icl-u-lg-mr--sm icl-u-xs-mr--xs"

func isCompany(n * html.Node) bool {
  return n.Type == html.ElementNode && n.Data == "div" &&
          len(n.Attr) == 1 && n.Attr[0].Val == company
}

//job role and location contained within <title></title>
func isTitle(n * html.Node) bool {
  return n.Type == html.ElementNode && n.Data == "title"
}

//depth first search across html
func traverse(n * html.Node, f func(n *html.Node) bool, ch chan string) bool {
    if f(n) {
      ch <- n.FirstChild.Data
      return true
    }

    for c := n.FirstChild; c != nil; c = c.NextSibling {
      if traverse(c, f, ch) {
        return true
      }
    }
    //corner case? <-ch
    return false
}
//provides more capability however does not fetch correct title
/*
func traverse(n * html.Node, f func(n *html.Node) bool, ch chan string) {
  var stack = []*html.Node{}
  stack = append(stack, n)

  for len(stack) > 0 {
    res := stack[len(stack) - 1]
    stack = stack[:len(stack) - 1]

    if f(res) {
      ch <- res.FirstChild.Data
      return
    }
    for c := res.FirstChild; c != nil; c = c.NextSibling {
      stack = append(stack, c)
    }
  }
}*/

//Parse request url page hmtl and return resp json
func fetch(url string) info {
  //get *Node for html page
  htm, err := http.Get(url)
  if err != nil {
    log.Fatal(err)
  }

  doc, _ := html.Parse(htm.Body)

  var is_t Comparator = isTitle
  var is_c Comparator = isCompany
  var title, company string

  ch_t := make(chan string)
  ch_c := make(chan string)
  //concurrently parse html
  go traverse(doc, is_t, ch_t)
  go traverse(doc, is_c, ch_c)

  for count := 0; count < chans; {
    select {
      case title = <-ch_t:
        count++
      case company = <-ch_c:
        count++
    }
  }

  //seperate/concatenate title & location fields
  fields := strings.Split(title, " - ")
  sz   :=  len(fields)
  //fill response fields
  return info{
            Title:    strings.Join(fields[:sz - offset], ""),
            Location: fields[sz - offset],
            Company:  company,
            Url:      url,
          }
}

//server side functions
func req_Handler(w http.ResponseWriter, r *http.Request) {
  //Recieve Request
  jsn, err := ioutil.ReadAll(r.Body)
  if err != nil {
    log.Fatal("Error reading the body", err)
  }

  jl := jsReq{}
  err = json.Unmarshal(jsn, &jl)
  if err != nil {
    log.Fatal(err)
  }

  //Prepare Response
  jr := jsResp{}

  //concurrency communication
  call := make(chan string, len(jl.Listings))
  resp := make(chan info, len(jl.Listings))

  //parallelism on multiprocessor machine
  cpus := runtime.NumCPU()
  var wg sync.WaitGroup
  for i := 0; i < cpus; i++ {
    wg.Add(1)
    go func(out chan<- info, in <-chan string) {
        for url := range in {
          out <- fetch(url)
        }
        wg.Done()
    }(resp, call)
  }

  //fetch job role, location, and company from url
  for _, url := range jl.Listings {
    call <- url
  }
  close(call)

  //sync responses
  wg.Wait()
  close(resp)
  for r := range resp {
    jr.Tlcu = append(jr.Tlcu, r)
  }

  fmt.Println("length")
  fmt.Println(len(jr.Tlcu))
  //Respond to client
  response,err := json.MarshalIndent(jr,"","  ")
  if err != nil {
    log.Fatal(err)
	}

  w.Header().Set(hKey, hVal)
  w.Write(response)
}

func server() {
  http.HandleFunc("/", req_Handler)
  http.ListenAndServe(":8080", nil)
}

//client side function
func client() {
  //open file
  file, err := os.Open(os.Args[input])
  out := os.Args[output]
  if err != nil {
    log.Fatal(err)
  }

  ourlist := jsReq{}

  //write req json formated list of url's
  scanner := bufio.NewScanner(file)
  for scanner.Scan() {
    ourlist.Listings = append(ourlist.Listings, scanner.Text())
  }
  file.Close()

  jl, err := json.Marshal(ourlist)
  if err != nil {
    log.Fatal(err)
  }

  //send POST request
  client := &http.Client{}

  req,err := http.NewRequest(reqMeth, reqUrl, bytes.NewBuffer(jl))
  req.Header.Set(hKey, hVal)
  //print req
  resp, err := client.Do(req)

  if err != nil {
    log.Fatal(err)
  }

  //Print Response & close
  body, err := ioutil.ReadAll(resp.Body)
  ioutil.WriteFile(out,body, 0644)
  resp.Body.Close()
}

func main() {
  go server()
  client()
}
