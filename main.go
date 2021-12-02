package main

import (
	"encoding/hex"
	"flag"
	"os"
	"path/filepath"
	"strconv"

	"github.com/davecgh/go-spew/spew"
	"github.com/devedge/imagehash"
	"github.com/steakknife/hamming"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type ImageFile struct {
	gorm.Model
	Path   string
	Hash   string
	SameTo int64
}

func getDb() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("data.db"), &gorm.Config{})
	if err != nil {
		spew.Dump(err)
		panic("failed to connect database")
	}
	return db
}

func scanDir(path string) {
	db := getDb()
	db.AutoMigrate(&ImageFile{})
	err := filepath.Walk(path, func(path string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}
		if f.IsDir() {
			return nil
		}
		spew.Dump(path)
		src, _ := imagehash.OpenImg(path)
		dhash, _ := imagehash.Dhash(src, 8)
		hash := hex.EncodeToString(dhash)
		db.Create(&ImageFile{Path: path, Hash: hash})
		return nil
	})
	if err != nil {
		spew.Dump(err)
		panic("scan dir error")
	}
}

func hashDistance(max int) {
	if max < 1 {
		max = 1
	}
	db := getDb()
	var files []ImageFile
	db.Order("id").Find(&files)
	for _, v := range files {
		if v.SameTo < 1 {
			for _, row := range files {
				if row.ID > v.ID && row.SameTo < 1 {
					b1, _ := hex.DecodeString(v.Hash)
					b2, _ := hex.DecodeString(row.Hash)
					distance := hamming.Bytes(b1, b2)
					spew.Dump(distance)
					if distance < max {
						row.SameTo = int64(v.ID)
						db.Save(&row)
					}
				}
			}
		}
	}
}

func main() {
	flag.Parse()
	command := flag.Arg(0)
	if command == "scan" {
		root := flag.Arg(1)
		if root != "" {
			scanDir(root)
		} else {
			panic("miss arg <root>")
		}
	} else if command == "distance" {
		max, _ := strconv.Atoi(flag.Arg(1))
		hashDistance(max)
	} else {
		panic("use as `self <command> <arg>`")
	}
}
