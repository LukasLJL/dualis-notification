package dualis

import (
	"log"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"strings"

	"github.com/yhat/scrape"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func (dualis *Dualis) updateModules() (updatedModules []Module) {
	for i, _ := range dualis.Semester {
		for j, _ := range dualis.Semester[i].Modules {
			var diffModule Module
			diffModule = dualis.Semester[i].Modules[j]

			newModule, _ := dualis.parseModule(&dualis.Semester[i].Modules[j])
			dualis.Semester[i].Modules[j] = *newModule

			if !reflect.DeepEqual(diffModule, dualis.Semester[i].Modules[j]) {
				updatedModules = append(updatedModules, dualis.Semester[i].Modules[j])
			}
		}
	}

	return updatedModules
}

func (dualis *Dualis) discoverModules(semester *Semester) {
	req, _ := http.NewRequest("GET", baseURL+semester.Url, nil)
	req.Header.Add("User-Agent", userAgent)
	resp, _ := dualis.Client.Do(req)

	root, _ := html.Parse(resp.Body)

	htmlModuleTable, _ := scrape.Find(root, scrape.ByClass("nb"))

	htmlModuleLinks := scrape.FindAll(htmlModuleTable, scrape.ByTag(atom.A))

	for _, htmlModuleLink := range htmlModuleLinks {
		module := Module{
			Url: scrape.Attr(htmlModuleLink, "href"),
		}

		semester.Modules = append(semester.Modules, module)
	}

	log.Printf("Discovered %v new modules in semester: %s", len(semester.Modules), semester.Name)
}

func (dualis *Dualis) discoverSemesters(url string) {
	req, _ := http.NewRequest("GET", baseURL+url, nil)
	req.Header.Add("User-Agent", userAgent)
	resp, _ := dualis.Client.Do(req)

	root, _ := html.Parse(resp.Body)

	semesterMatcher := func(n *html.Node) bool {
		return n.DataAtom == atom.Option
	}

	htmlSemesterSelect, _ := scrape.Find(root, scrape.ById("semester"))

	semesterBaseUrl := dualis.buildSemesterUrl(scrape.Attr(htmlSemesterSelect, "onchange"))

	for _, htmlSemester := range scrape.FindAllNested(htmlSemesterSelect, semesterMatcher) {
		semester := Semester{
			Name: scrape.Text(htmlSemester),
			Url:  semesterBaseUrl + scrape.Attr(htmlSemester, "value"),
		}
		dualis.Semester = append(dualis.Semester, semester)

		log.Println("Discovered new semester:", semester.Name)
	}
}

func (dualis *Dualis) buildSemesterUrl(dirt string) (url string) {
	regex := regexp.MustCompile(`(?:')(.*?)(?:')`)

	var params []string

	for _, match := range regex.FindAllStringSubmatch(dirt, -1) {
		params = append(params, match[1])
	}

	url = params[0] + "?APPNAME=" + params[1] + "&PRGNAME=" + params[2] + "&ARGUMENTS=-N" + params[3] + ",-N" + params[4] + "," + params[5]

	return url
}

func (module *Module) equal(b *Module) bool {
	return reflect.DeepEqual(module, b)
}

func (dualis *Dualis) initStructs(homeUrl *url.URL) {
	req, _ := http.NewRequest("GET", baseURL+homeUrl.String(), nil)
	req.Header.Add("User-Agent", userAgent)
	resp, _ := dualis.Client.Do(req)

	navElementMatcher := func(n *html.Node) bool {
		return scrape.Attr(n, "title") == "Prüfungsergebnisse"
	}

	root, _ := html.Parse(resp.Body)
	htmlNavElement, _ := scrape.Find(root, navElementMatcher)
	htmlNavLink, _ := scrape.Find(htmlNavElement, scrape.ByTag(atom.A))

	dualis.discoverSemesters(scrape.Attr(htmlNavLink, "href"))

	for i, semester := range dualis.Semester {
		dualis.discoverModules(&semester)
		dualis.Semester[i] = semester
	}
}

func (dualis *Dualis) sessionCookie(resp *http.Response) (cookie string, ok bool) {
	//cookie fix
	c := resp.Header["Set-Cookie"]
	for i, v := range c {
		name, value, ok := strings.Cut(v, "=")
		if !ok {
			continue
		}
		if nt := strings.TrimSpace(name); name != nt {
			c[i] = nt + "=" + value
		}
	}
	resp.Header["Set-Cookie"] = c

	if len(resp.Cookies()) > 0 && resp.Cookies()[0].Name == "cnsc" {
		log.Println("Session cookie created.")
		return resp.Cookies()[0].Value, true
	} else {
		log.Println("Session cookie not created.")
		return resp.Cookies()[0].Value, false
	}
}

func (dualis *Dualis) cleanRefreshURL(dirt string) (cleanURL *url.URL, ok bool) {
	regex, _ := regexp.Compile(`\bURL=(.*)`)

	// This is probably the worst way of finding the correct string - TODO
	match := regex.FindStringSubmatch(dirt)

	cleanURL, error := url.Parse(match[1])

	if error != nil {
		var u *url.URL
		return u, false
	}

	return cleanURL, true
}

func (dualis *Dualis) parseModule(module *Module) (mod *Module, ok bool) {
	url := module.Url

	req, _ := http.NewRequest("GET", baseURL+url, nil)
	req.Header.Add("User-Agent", userAgent)
	resp, _ := dualis.Client.Do(req)

	root, _ := html.Parse(resp.Body)

	rowMatcher := func(n *html.Node) bool {
		return n.DataAtom == atom.Tr
	}

	columnMatcher := func(n *html.Node) bool {
		return n.DataAtom == atom.Td
	}

	moduleNameMatcher := func(n *html.Node) bool {
		return n.DataAtom == atom.H1
	}

	htmlRows := scrape.FindAll(root, rowMatcher)

	htmlModuleName, _ := scrape.Find(root, moduleNameMatcher)

	module.Name = strings.Replace(scrape.Text(htmlModuleName), "\n", "", -1)

	// Reset module
	module.Attempts = []Attempt{}

	processingEvent := false

ProcessRows:
	for _, row := range htmlRows {
		htmlColumns := scrape.FindAll(row, columnMatcher)
		//fmt.Println(scrape.Text(row))

		switch scrape.Attr(htmlColumns[0], "class") {
		case "level01":
			log.Printf("┌ Attempt: %s\n", scrape.Text(htmlColumns[0]))

			attempt := Attempt{
				Label: scrape.Text(htmlColumns[0]),
			}
			module.Attempts = append(module.Attempts, attempt)

		case "level02":
			if processingEvent && len(htmlColumns) > 1 && scrape.Attr(htmlColumns[1], "class") != "level02 level02_chkbox_workaround_mpa" {
				log.Printf("└ Event result: %s\n", scrape.Text(htmlColumns[3]))

				processingEvent = false

				// Intoducing these variables to combat heavy nesting of slice keys
				//attempt := module.Attempts[len(module.Attempts)-1]
				//event := attempt.Events[len(attempt.Events)-1]

				//event.Grade = scrape.Text(htmlColumns[3])
				module.Attempts[len(module.Attempts)-1].Events[len(module.Attempts[len(module.Attempts)-1].Events)-1].Grade = scrape.Text(htmlColumns[3])
			} else {
				log.Printf("├┬ New event: %s\n", scrape.Text(htmlColumns[0]))

				processingEvent = true

				event := Event{
					Name: scrape.Text(htmlColumns[0]),
				}
				module.Attempts[len(module.Attempts)-1].Events = append(module.Attempts[len(module.Attempts)-1].Events, event)
			}

		case "tbdata":
			log.Printf("│├─ Exam: %s\n", scrape.Text(htmlColumns[1]))

			// Intoducing these variables to combat heavy nesting of slice keys
			//attempt := module.Attempts[len(module.Attempts)-1]
			//event := attempt.Events[len(attempt.Events)-1]

			exam := Exam{
				Semester: scrape.Text(htmlColumns[0]),
				Name:     scrape.Text(htmlColumns[1]),
				Grade:    scrape.Text(htmlColumns[3]),
			}

			//event.Exams = append(event.Exams, exam)
			module.Attempts[len(module.Attempts)-1].Events[len(module.Attempts[len(module.Attempts)-1].Events)-1].Exams = append(module.Attempts[len(module.Attempts)-1].Events[len(module.Attempts[len(module.Attempts)-1].Events)-1].Exams, exam)

		case "tbhead":
			if scrape.Text(htmlColumns[0]) == "Pflichtbereich" {
				break ProcessRows
			}
		}

	}
	processingEvent = false

	return module, true
}
