package graph

import (
	"encoding/gob"
	"encoding/json"
	"os"
)

type Storage interface {
	Save(g *Graph) error
	Load() (*Graph, error)
}

func NewStorage(path string, format string) Storage {
	switch format {
	case "binary":
		return &BinaryStorage{filename: path}
	default:
		return &JsonStorage{filename: path}
	}
}

type BinaryStorage struct {
	filename string
}

func (bs *BinaryStorage) Save(g *Graph) error {
	file, err := os.Create(bs.filename)
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := gob.NewEncoder(file)
	return encoder.Encode(g)
}

func (bs *BinaryStorage) Load() (*Graph, error) {
	file, err := os.Open(bs.filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var g Graph
	decoder := gob.NewDecoder(file)
	if err := decoder.Decode(&g); err != nil {
		return nil, err
	}
	return &g, nil
}

type JsonStorage struct {
	filename string
}

func (js *JsonStorage) Save(g *Graph) error {
	file, err := os.Create(js.filename)
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	return encoder.Encode(g)
}

func (js *JsonStorage) Load() (*Graph, error) {
	file, err := os.Open(js.filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var g Graph
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&g); err != nil {
		return nil, err
	}
	return &g, nil
}
