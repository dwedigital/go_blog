package main

import (
	"errors"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"fmt"
)

type Page struct {
	Title string
	Body []byte
}



// cache templates. Any new templates need to be added in this list
var templates = template.Must(template.ParseFiles("./templates/edit.html", "./templates/view.html", "./templates/index.html"))

// regex to check only allowed paths are used
var validPath = regexp.MustCompile("^/(edit|save|post)/([a-zA-Z0-9]+)$")

func init(){
	// create the posts directory if it doesn't exist
	os.Mkdir("./posts", 0777)

}

func main(){
	http.HandleFunc("/post/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	http.HandleFunc("/", indexHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// save page - p is a Page pointer and is a receiever not a paramaeter of the function
func (p *Page) save() error {
	filename := p.Title + ".txt"
	return os.WriteFile(filename, p.Body, 0600)
}


// load the page
func loadPage(title string) (*Page, error) {
	filename := title + ".txt"
	body, err := os.ReadFile("./posts/"+filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func getTitle (w http.ResponseWriter, r *http.Request) (string, error){
	m := validPath.FindStringSubmatch(r.URL.Path)
	if m == nil {
		http.NotFound(w, r)
		return "", errors.New("Invalid Page Title")
	}
	return m[2], nil
}

func makeHandler (fn func (http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2])
	}
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string){

	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	renderTemaplate(w, "view", p)
}

func editHandler (w http.ResponseWriter, r *http.Request, title string){
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemaplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string){
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func renderTemaplate (w http.ResponseWriter, tmpl string, p *Page){
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}
 func indexHandler (w http.ResponseWriter, r *http.Request){
	posts, err := ioutil.ReadDir("./posts")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	links := make([]string, len(posts))


	for i := 0; i < len(posts); i++ {
		link := strings.TrimSuffix(posts[i].Name(), ".txt")
		
		// append link to links array
		links[i] = link
	}

	// print each item in links array
	for i := 0; i < len(links); i++ {
		fmt.Println(links[i])
	}
	
	templates.ExecuteTemplate(w, "index.html", links)
 }