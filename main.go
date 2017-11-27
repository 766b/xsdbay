package main

import (
	"bytes"
	"encoding/xml"
	"flag"
	"fmt"
	"go/format"
	"hash/fnv"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

type Xyer interface {
	GetName() string
	GetRelated() Xyer
	GetType() Type
	GetElements() []Xyer
	GoLine() string
	Generate()
	Validator(callName, path string)
	DeepValidator(callName, path string) bool
	Setter(typeName string)
}

type fileExt byte

const (
	extXSD fileExt = iota + 1
	extWSDL
)

var (
	// checkMode      = flag.Int("check", 0, "Check for changes only")
	inputFilePath  = flag.String("i", "", "Input file")
	exportElements = flag.String("e", "", "Elements to be exported (comma separated)")
	outputFilePath = flag.String("o", "", "Output Go file (Default: ebaysvc_####.go)")

	apiVersion = flag.String("apiver", "", "API Version")
	latestXSD  = flag.Bool("latest", false, "Download latest version")
	cacheXSD   = flag.Bool("cache-xsd", false, "Cache downloaded file")
	onlineXSD  = flag.String("download", "http://developer.ebay.com/webservices/latest/ebaysvc.xsd", "XSD link")

	onlineMask string = "http://developer.ebay.com/webservices/%d/ebaysvc.xsd"

	fileType fileExt
	file     string
	// pkgName          string
	wsdlSc           definitions
	xsdSc            schema
	exportedElements []string

	Types     map[string]buffer = make(map[string]buffer)
	Calls     map[string]buffer = make(map[string]buffer)
	Enums     map[string]buffer = make(map[string]buffer)
	Funcs     map[string]buffer = make(map[string]buffer)
	Validator map[string]buffer = make(map[string]buffer)
)

type buffer struct {
	*bytes.Buffer
}

func (s buffer) Sprintf(str string, args ...interface{}) {
	s.WriteString(fmt.Sprintf(str, args...))
}

func NewBuffer() buffer {
	return buffer{bytes.NewBuffer(nil)}
}

func readInputFile() {
	var data []byte
	var err error
	if *latestXSD {
		_, file = path.Split(*onlineXSD)
		log.Println("Downloading latest file from ", *onlineXSD)
		resp, err := http.Get(*onlineXSD)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()
		data, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		fileType = extXSD
	} else {
		_, file = path.Split(*inputFilePath)
		log.Println("Reading file from ", *inputFilePath)
		if strings.HasSuffix(*inputFilePath, ".xsd") {
			fileType = extXSD
		} else if strings.HasSuffix(*inputFilePath, ".wsdl") {
			fileType = extWSDL
		}
		data, err = ioutil.ReadFile(*inputFilePath)
		if err != nil {
			log.Fatal(err)
		}
	}

	switch fileType {
	case 1:
		err = xml.Unmarshal(data, &xsdSc)
	case 2:
		err = xml.Unmarshal(data, &wsdlSc)
	}
	if err != nil {
		log.Fatal(err)
	}

	if fileType == extWSDL {
		*apiVersion = wsdlSc.Service.Documentation.Version
	}
	if *apiVersion == "" && fileType == extXSD {
		Vers := regexp.MustCompile(`<!-- Version (\d{4}) -->`).FindStringSubmatch(string(data[:80]))
		if len(Vers) == 2 {
			*apiVersion = Vers[1]
			log.Printf("API Version: %s", *apiVersion)
		}
	}

	if *apiVersion == "" {
		log.Fatalf("could identify API version. Use flag -apiver with right version number")
	}

	if *cacheXSD {
		ioutil.WriteFile(fmt.Sprintf("%s_%s", *apiVersion, file), data, 0644)
	}
}

func main() {
	start := time.Now()
	flag.Parse()

	readInputFile()

	var filePath string = *outputFilePath
	if filePath == "" {
		filePath = fmt.Sprintf("./%s_%s.go", file, *apiVersion)
	}

	fw, err := os.Create(filePath)
	if err != nil {
		log.Fatal(err)
	}

	if *exportElements == "" { //|| *checkMode != 0
		loadAllCalls()
	} else {
		exportedElements = strings.Split(strings.Replace(*exportElements, " ", "", -1), ",")
	}

	for _, e := range exportedElements {
		log.Printf("Call: %s", e)
		FromRequest(e)
		FromResponse(e)
	}

	// if *checkMode > 0 {
	// 	compare()
	// 	return
	// }

	fo := bytes.NewBufferString(fmt.Sprintf(templateEbaySVC, *apiVersion))
	fo.WriteString(templateNulls)

	fo.WriteString("type Request struct {\r\n")
	for _, val := range exportedElements {
		fo.WriteString(val + "Request " + val + "RequestType\r\n")
	}
	fo.WriteString("}\r\n\r\n")

	for _, v := range []map[string]buffer{Types, Calls, Enums, Funcs} {
		for _, val := range v {
			// fo.WriteString(fmt.Sprintf("//go:generate xsdbay -check=%d -latest -e=%s\r\n", hash(val.String()), k))
			fo.Write(val.Bytes())
			fo.WriteString("\r\n")
		}
	}

	for k, val := range Validator {
		if val.Len() == 0 {
			continue
		}
		fo.WriteString(requester(k))
		fo.WriteString(xmlEncoder(k))
		fo.WriteString(xmlMarshaler(k))
		// fo.WriteString(fmt.Sprintf("//go:generate xsdbay -check=%d -latest -v -e=%s\r\n", hash(val.String()), k))
		fo.WriteString(validator(k, val.String()))
	}

	fw.Write(formatCode(fo.Bytes()))
	fw.Close()
	log.Printf("Completed in %s.", time.Since(start))
}

// func compare() {
// 	for k, v := range Types {
// 		if k == *exportElements {
// 			if hash(v.String()) != uint32(*checkMode) {
// 				log.Printf("%s: Hashes do not match", k)
// 			} else {
// 				log.Printf("%s: OK", k)
// 			}
// 		}
// 	}
// }

func loadAllCalls() {
	for _, e := range getSchema().Element {
		if strings.HasSuffix(e.Name, "Request") {
			exportedElements = append(exportedElements, strings.TrimSuffix(e.Name, "Request"))
		}
	}
}

func formatCode(b []byte) []byte {
	source, err := format.Source(b)
	if err != nil {
		errLine, err2 := strconv.Atoi(strings.Split(fmt.Sprintf("%s", err), ":")[0])
		lines := strings.Split(string(b), "\n")
		for i, line := range lines {
			if err2 == nil {
				if i+1 > errLine-5 && i+1 < errLine+5 {
					pnt := " "
					if errLine == i+1 {
						pnt = `█`
					}
					fmt.Printf("% -10d |%s| %s\n", i+1, pnt, line)
				}
			} else {
				fmt.Printf("% -10d | %s\n", i+1, line)
			}

		}
		log.Fatalf("\n\nFailed to format source code. Error: %s", err)
	}
	return source
}

func UpperFirstLetter(x string) string {
	return strings.ToUpper(x[0:1]) + x[1:]
}

func LowerFirstLetter(x string) string {
	return strings.ToLower(x[0:1]) + x[1:]
}

func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func FromRequest(name string) {
	for _, e := range getSchema().Element {
		if e.Name == name+"Request" {
			x, _ := FindComplex(e.Type.String())
			x.Generate()
			x.Validator(name, "")
			x.Setter("")
			return
		}
	}
	log.Fatal("could not find element: " + name + "Request")
}

func FromResponse(name string) {
	for _, e := range getSchema().Element {
		if e.Name == name+"Response" {
			x, _ := FindComplex(e.Type.String())
			x.Generate()
			x.Setter("")
			return
		}
	}
	log.Fatal("could not find element: " + name + "Response")
}

func Find(name string) Xyer {
	if b, ok := FindComplex(name); ok {
		return b
	}
	if b, ok := FindSimple(name); ok {
		return b
	}
	panic("Could not find " + name)
	return nil
}

func getSchema() *schema {
	switch fileType {
	case extXSD:
		return &xsdSc
	case extWSDL:
		return &wsdlSc.Types.Schema
	}
	return nil
}

func FindComplex(name string) (*complexType, bool) {
	for _, x := range getSchema().ComplexType {
		if x.Name == name {
			return &x, true
		}
	}
	return nil, false
}

func FindSimple(name string) (*simpleType, bool) {
	for _, x := range getSchema().SimpleType {
		if x.Name == name {
			return &x, true
		}
	}
	return nil, false
}

// Credit: https://gist.github.com/elwinar/14e1e897fdbe4d3432e1
func ToSnake(in string) string {
	runes := []rune(in)
	length := len(runes)

	var out []rune
	for i := 0; i < length; i++ {
		if i > 0 && unicode.IsUpper(runes[i]) && ((i+1 < length && unicode.IsLower(runes[i+1])) || unicode.IsLower(runes[i-1])) {
			out = append(out, '_')
		}
		out = append(out, unicode.ToLower(runes[i]))
	}

	return string(out)
}
