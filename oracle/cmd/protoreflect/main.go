package main

import (
	"fmt"
	"log"

	consulApi "github.com/hashicorp/consul/api"
	"github.com/jhump/protoreflect/desc/protoparse"
	"google.golang.org/protobuf/types/descriptorpb"
)

func parseProtoFile() error {
	fileNames := []string{"service/slark/http.proto"}
	importPaths := []string{"proto/api"}
	fileNames, err := protoparse.ResolveFilenames(importPaths, fileNames...)
	if err != nil {
		return err
	}
	p := protoparse.Parser{
		ImportPaths:           importPaths,
		InferImportPaths:      len(importPaths) == 0,
		IncludeSourceCodeInfo: true,
	}
	fds, err := p.ParseFiles(fileNames...)
	if err != nil {
		return err
	}
	for i, fd := range fds {
		log.Printf("name of fd[%d]: %s\n", i, fd.GetName())
		log.Printf("fully qualified name of fd[%d]: %s\n", i, fd.GetFullyQualifiedName())
		log.Printf("package of fd[%d]: %s\n", i, fd.GetPackage())
		sds := fd.GetServices()
		for j, sd := range sds {
			log.Printf("name of sd[%d]: %s\n", j, sd.GetName())
			mds := sd.GetMethods()
			for k, md := range mds {
				log.Printf("name of md[%d]: %s\n", k, md.GetName())
				symbol := md.GetFullyQualifiedName()
				log.Printf("fully qualified name of md[%d]: %s\n", k, symbol)
				d := fd.FindSymbol(symbol)
				log.Printf("name of fd.FindSymbol(%s): %s\n", symbol, d.GetName())
				output := md.GetOutputType()
				log.Printf("output type of md[%d]: %s\n", k, output.String())
				fileds := output.GetFields()
				for l, field := range fileds {
					fieldType := field.GetType()
					switch fieldType {
					case descriptorpb.FieldDescriptorProto_TYPE_MESSAGE:
						nestFields := field.GetMessageType().GetFields()
						log.Printf("fully qualified name of field of md[%d] field [%d]: %s\n", k, l, field.GetMessageType().GetFullyQualifiedName())
						for _, nf := range nestFields {
							log.Printf("json name of nest field of md[%d] field [%d]: %s\n", k, l, nf.GetJSONName())
							log.Printf("IsRepeated of nest field of md[%d] field [%d]: %v\n", k, l, nf.IsRepeated())
							log.Printf("IsRepeated of nest field of md[%d] field [%d]: %s\n", k, l, nf.GetLabel())
						}
					}
					log.Printf(" json name of field of md[%d]: %s\n", k, field.GetJSONName())
				}
				// mdops := md.GetMethodOptions()
				// log.Printf("mdops of md[%d]: %s\n", k, mdops.String())
			}
		}
	}
	return nil
}

func main() {
	// parseProtoFile()
	// fp := "proto/api/service/slark/http.proto"
	// data, err := os.ReadFile(fp)
	// if err != nil {
	// 	log.Fatalln(err)
	// }
	// re := regexp.MustCompile(`post\s*:\s*"(.+)"`)
	// matrix := re.FindAllStringSubmatch(string(data), -1)
	// for _, slice := range matrix {
	// 	fmt.Println(slice[1])
	// }

	config := consulApi.DefaultConfig()
	client, err := consulApi.NewClient(config)
	if err != nil {
		log.Fatalln(err)
	}
	services, _, err := client.Catalog().Services(nil)
	if err != nil {
		log.Fatalln(err)
	}
	for srv := range services {
		// health check
		entries, _, err := client.Health().Service(srv, "", true, nil)
		if err != nil {
			log.Fatalln(err)
		}
		for _, entry := range entries {
			fmt.Printf("Service: %s, Node: %s\n",
				entry.Service.Service, entry.Node.Node)
			for idx, check := range entry.Checks {
				fmt.Printf("index [%d] check.Status: %s\n",
					idx, check.Status)
			}
		}
		if entries[0].Checks[0].Status != "passing" {
			log.Fatalln(err)
		}
	}
}
