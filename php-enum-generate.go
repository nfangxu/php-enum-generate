package php_enum_generate

import (
    "encoding/json"
    "flag"
    "fmt"
    "io"
    "io/ioutil"
    "os"
    "path/filepath"
    "strings"
    "unicode"
)

var docs = flag.String("doc", "./grpc-doc.json", "proto 生成的 json 文件")

// doc json struct
type EnumValue struct {
    Name        string `json:"name"`
    Number      string `json:"number"`
    Description string `json:"description"`
}

type Enum struct {
    Name        string      `json:"name"`
    LongName    string      `json:"longName"`
    FullName    string      `json:"fullName"`
    Description string      `json:"description"`
    Values      []EnumValue `json:"values"`
}

type File struct {
    Name          string        `json:"name"`
    Description   string        `json:"description"`
    Package       string        `json:"package"`
    HasEnums      bool          `json:"hasEnums"`
    HasExtensions bool          `json:"hasExtensions"`
    HasMessages   bool          `json:"hasMessages"`
    HasServices   bool          `json:"hasServices"`
    Enums         []Enum        `json:"enums"`
    Extensions    []interface{} `json:"extensions"`
    Messages      []interface{} `json:"messages"`
    Services      []interface{} `json:"services"`
}

type Doc struct {
    Files            []File `json:"files"`
    ScalarValueTypes []struct {
        ProtoType  string `json:"protoType"`
        Notes      string `json:"notes"`
        CppType    string `json:"cppType"`
        CsType     string `json:"csType"`
        GoType     string `json:"goType"`
        JavaType   string `json:"javaType"`
        PhpType    string `json:"phpType"`
        PythonType string `json:"pythonType"`
        RubyType   string `json:"rubyType"`
    } `json:"scalarValueTypes"`
}

func main() {
    flag.Parse()

    fmt.Printf("use doc file: [%s]\n", *docs)

    file, err := os.Open(*docs)

    if err != nil {
        panic(err)
    }

    content, err := ioutil.ReadAll(file)

    data := json.NewDecoder(strings.NewReader(string(content)))

    var item Doc
    if err := data.Decode(&item); err == io.EOF {
        panic(err)
    } else if err != nil {
        panic(err)
    }

    for i := 0; i < len(item.Files); i++ {
        file := item.Files[i]
        if file.HasEnums {
            for i := 0; i < len(file.Enums); i++ {
                enum := file.Enums[i]
                build(stub(), ucFirst(enum.FullName), enum.Values)
            }
        }
    }
}

func stub() string {
    return `
<?php
namespace DummyNamespace;

class DummyClass
{
    use \App\Enums\Enum;

    DummyConstants
}

`
}

func build(stub string, name string, values []EnumValue) {
    classname := strReplace(".", "\\", name)
    _ = os.MkdirAll(filepath.Dir(getFilename(classname)), 0777)
    c := ""
    for i := 0; i < len(values); i++ {
        c = c + getConstant(values[i])
    }
    stub = strReplace("DummyNamespace", getNamespace(classname), stub)
    stub = strReplace("DummyClass", getClassname(classname), stub)
    stub = strReplace("DummyConstants", strings.TrimSpace(c), stub)
    f, _ := os.Create(getFilename(classname))
    _, _ = f.WriteString(strings.TrimSpace(stub) + "\n")
    fmt.Println(fmt.Sprintf("[%s] was created.", classname))
}

func getNamespace(classname string) string {
    prefix := "ProtoCenter\\"
    if classname != getClassname(classname) {
        return prefix + strReplace("\\"+getClassname(classname), "", classname)
    } else {
        return prefix
    }
}

func getClassname(classname string) string {
    s := strings.Split(classname, "\\")
    return s[len(s)-1]
}

func getConstant(v EnumValue) string {
    return fmt.Sprintf("    const %s = %s, __%s = \"%s\"; \n", v.Name, v.Number, v.Name, v.Description)
}

func getFilename(classname string) string {
    return fmt.Sprintf("./src/%s.php", strReplace("\\", "/", classname))
}

func ucFirst(str string) string {
    for _, v := range str {
        u := string(unicode.ToUpper(v))
        return u + str[len(u):]
    }
    return ""
}

func strReplace(search, replace, subject string) string {
    return strings.ReplaceAll(subject, search, replace)
}
