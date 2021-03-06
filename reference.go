package main

import (
	"archive/zip"
	"encoding/json"

	"golang.org/x/tools/godoc/vfs"
	"golang.org/x/tools/godoc/vfs/zipfs"
)

// Reference stores information related to all programming language references
// available to the user.
type Reference struct {
	Data       []ReferenceData
	FileSystem vfs.FileSystem
}

// ReferenceData stores information related to a single reference.
type ReferenceData struct {
	Name  string `json:"name"`
	Title string `json:"title"`
	Index string `json:"index"`
}

// OpenReference is used to import information about all references stored at
// the specified location.
func OpenReference(location string) (*Reference, error) {
	rc, err := zip.OpenReader(location)
	if err != nil {
		return nil, err
	}

	fs := zipfs.New(rc, "reference_zipfs")
	info, err := vfs.ReadFile(fs, "/info.json")
	if err != nil {
		return nil, err
	}

	var data []ReferenceData
	err = json.Unmarshal(info, &data)
	if err != nil {
		return nil, err
	}

	return &Reference{Data: data, FileSystem: fs}, nil
}
