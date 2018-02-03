package spacer

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/mndrix/tap-go"
	blackfriday "gopkg.in/russross/blackfriday.v2"
)

func RunTest(host string, testPath string) error {
	filepath.Walk(testPath, func(path string, info os.FileInfo, err error) error {

		if info.IsDir() {
			return nil
		}

		t, err := NewTest(host, path)
		if err != nil {
			log.Fatal(err)
		}
		err = t.Run()
		if err != nil {
			log.Fatal(err)
		}
		return nil
	})

	return nil
}

type Runner struct {
}

type Test struct {
	Name         string
	Description  string
	Method       string
	Url          *url.URL
	Header       map[string]string
	Body         string
	ExpectCode   int
	ExpectBody   string
	ExpectHeader string
}

func (t *Test) Run() error {
	plan := tap.New()
	plan.Diagnosticf("%s", t.Url.Path)
	client := &http.Client{}
	req, err := http.NewRequest(
		t.Method,
		t.Url.String(),
		bytes.NewReader([]byte(t.Body)),
	)

	for k, v := range t.Header {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	plan.Ok(resp.StatusCode == t.ExpectCode, "status code")
	if resp.StatusCode != t.ExpectCode {
		plan.Diagnosticf("Expect %d, got %d", t.ExpectCode, resp.StatusCode)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	bodyStr := strings.TrimSpace(string(body))

	plan.Ok(bodyStr == t.ExpectBody, "body")
	if bodyStr != t.ExpectBody {
		plan.Diagnosticf("Expect body %s, got %s", t.ExpectBody, bodyStr)
	}
	plan.AutoPlan()

	return nil
}

func NewTest(host string, path string) (Test, error) {
	t := Test{}
	md := blackfriday.New()

	b, err := ioutil.ReadFile(path)
	if err != nil {
		return t, err
	}

	ast := md.Parse(b)
	nextIsDesc := false
	nextIsExpectBody := false
	ast.Walk(func(node *blackfriday.Node, entering bool) blackfriday.WalkStatus {
		if !entering {
			return 0
		}

		if name, ok := MatchHeader(node, 1, "test"); ok {
			t.Name = name
			nextIsDesc = true
			return 0
		}

		if node.Type == blackfriday.Paragraph && node.Level == 0 && nextIsDesc {
			text := node.FirstChild
			t.Description = string(text.Literal)
			nextIsDesc = false
			return 0
		}

		if node.Type == blackfriday.Paragraph && node.Level == 0 && nextIsExpectBody {
			text := firstChildText(node)
			r := regexp.MustCompile(`^json\n`)
			text = strings.TrimSpace(r.ReplaceAllString(text, ""))
			t.ExpectBody = string(text)
			nextIsExpectBody = false
			return 0
		}

		if req, ok := MatchHeader(node, 2, "request"); ok {
			req := strings.Split(req, " ")
			method := req[0]
			path := req[1]

			url, err := url.Parse(fmt.Sprintf("%s/%s", host, path))
			if err != nil {
				log.Fatal(err)
			}
			t.Method = method
			t.Url = url
			return 0
		}

		if _, ok := MatchHeader(node, 2, "expect body"); ok {
			nextIsExpectBody = true
			return 0
		}

		if expectCode, ok := MatchHeader(node, 2, "expect"); ok {
			code, err := strconv.Atoi(expectCode)
			if err != nil {
				log.Fatal(err)
			}
			t.ExpectCode = code
			return 0
		}

		return 0
	})

	return t, nil
}

func MatchHeader(node *blackfriday.Node, level int, prefix string) (string, bool) {
	if node.Type == blackfriday.Heading && node.Level == level {
		text := node.FirstChild
		if text != nil && strings.HasPrefix(string(text.Literal), prefix) {
			r := regexp.MustCompile(fmt.Sprintf(`^%s\s`, prefix))
			cdr := r.ReplaceAllString(string(text.Literal), "")
			return cdr, true
		}
	}

	return "", false
}

func firstChildText(node *blackfriday.Node) string {
	c := node.FirstChild
	for string(c.Literal) == "" {
		c = c.Next
		if c == nil {
			return ""
		}
	}
	return string(c.Literal)
}
