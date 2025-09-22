package simpleproto

import (
	"encoding/base64"
	"os"
	"path/filepath"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

func ParseProtoFiles(filenames ...string) ([]*desc.FileDescriptor, error) {
	importPaths := []string{"proto/api"}
	filenames, err := protoparse.ResolveFilenames(importPaths, filenames...)
	if err != nil {
		return nil, err
	}
	p := protoparse.Parser{
		ImportPaths:           importPaths,
		InferImportPaths:      len(importPaths) == 0,
		IncludeSourceCodeInfo: true,
	}
	return p.ParseFiles(filenames...)
}

func FileDescriptorProto(filename string) ([]byte, error) {
	fds, err := ParseProtoFiles(filename)
	if err != nil {
		return nil, err
	}
	return proto.Marshal(fds[0].AsFileDescriptorProto())
}

func Base64FileDescriptorProto(filename string, data []byte) (string, error) {
	if len(data) > 0 {
		fileFullpath := "proto/api/" + filename
		// check dir
		if err := os.MkdirAll(filepath.Dir(fileFullpath), 0750); err != nil && !os.IsExist(err) {
			return "", err
		}
		// write file
		if err := os.WriteFile(fileFullpath, data, 0666); err != nil {
			return "", err
		}
		defer func() {
			os.Remove(fileFullpath)
		}()
	}
	data, err := FileDescriptorProto(filename)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

func FileDescriptorProtoFromBase64(b64fdd string) (*descriptorpb.FileDescriptorProto, error) {
	data, err := base64.StdEncoding.DecodeString(b64fdd)
	if err != nil {
		return nil, err
	}
	var res descriptorpb.FileDescriptorProto
	if err := proto.Unmarshal(data, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

type ServiceCache struct {
	ID             int64                `json:"id"`
	Name           string               `json:"name"`
	ApplicationID  int64                `json:"applicationID"`
	Application    string               `json:"application"`
	URL            string               `json:"url"`
	PathPrefix     string               `json:"pathPrefix"`
	ProtoFileMd5   string               `json:"protoFileMd5"`
	CreatedAt      int64                `json:"createdAt"`
	FileDescriptor *desc.FileDescriptor `json:"-"`
}
