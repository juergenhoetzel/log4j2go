package log4j

import (
	"archive/zip"
	"bytes"
	"crypto/md5"
	"encoding/xml"
	"fmt"
	"github.com/juergenhoetzel/log4j2go/internal/filedata"
	"io"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strings"
)

func isLog4jEntry(f *zip.File) bool {
	return strings.Contains(f.Name, "org/apache/logging/log4j/core") && strings.HasSuffix(f.Name, ".class") && f.UncompressedSize > 0
}

func Log4jHash(files []*zip.File) (string, bool) {
	checksum := md5.New()
	sortedFiles := []*zip.File{}
	isLog4j := false
	for _, f := range files {
		if isLog4jEntry(f) {
			isLog4j = true
			sortedFiles = append(sortedFiles, f)
		}
	}
	sort.Slice(sortedFiles, func(i, j int) bool {
		return sortedFiles[i].Name < sortedFiles[j].Name
	})
	for _, f := range sortedFiles {
		rdr, err := f.Open()
		if err != nil {
			log.Fatal(err)
		}
		io.Copy(checksum, rdr)
	}
	if isLog4j {
		return fmt.Sprintf("%x", checksum.Sum(nil)), true
	}
	return "", false
}

func JarVersion(r []*zip.File) string {
	// Fastest method: pom.xml lookup
	for _, f := range r {
		if strings.HasSuffix(f.Name, "log4j-core/pom.xml") {
			xrdr, err := f.Open()
			if err != nil {
				log.Fatalf("Faild to open %q: %v", f.Name, err)
			}
			defer xrdr.Close()
			decoder := xml.NewDecoder(xrdr)

			var groupId, artifactId, version bool
			// assumption order: groupId -> artifactId -> version
			for {
				t, err := decoder.Token()
				if err != nil {
					if err == io.EOF {
						break
					}
					log.Fatalf("Error decoding %q, err: %v", f.Name, err)
				}
				if se, ok := t.(xml.StartElement); ok {
					if se.Name.Local == "groupId" {
						groupId = true
					}
					if se.Name.Local == "artifactId" {
						artifactId = true
					}
					if se.Name.Local == "version" {
						version = true
					}
					continue
				}
				if groupId && artifactId && version {
					if cd, ok := t.(xml.CharData); ok {
						return fmt.Sprintf("log4j-core-%s", cd)
					}
				}
			}

			log.Fatalf("Found POM: %v:%v", f.Name, "!")
		}
	}
	// Fallback to hashes
	hash, ok := Log4jHash(r)
	if ok {
		if jarFile, ok := filedata.Filehashes[hash]; ok {
			return jarFile
		}
	}
	return ""
}

func CheckFile(zipname, zipfile string) string {
	// check for recursion
	rdr, err := zip.OpenReader(zipfile)
	if err != nil {
		// Second try: search for embedded zip
		ziperr := err
		magic := []byte{0x50, 0x4B, 0x3, 0x4}
		f, err := os.Open(zipfile)
		if err != nil {
			log.Fatalf("Failed to open %q: %v", zipfile, err)
		}
		pos := 0
		buffer := make([]byte, 1024)
		for {
			n, err := f.Read(buffer)
			if err != nil {
				if err != io.EOF {
					log.Fatalf("Failed to read %q: %s\n", zipfile, err)
				} else {
					log.Printf("%q does not contain a zip signature", f.Name())
					return ""
				}
			}
			if n > 0 {
				index := bytes.Index(buffer, magic)
				if index != -1 {
					_, err := f.Seek(int64(pos+index), 0)
					if err != nil {
						log.Fatalf("Seek to %d in %q faild: ", pos+index, err)
					}
					tf, err := ioutil.TempFile(os.TempDir(), "log4j2go-*.jar")
					io.Copy(tf, f)
					defer os.Remove(tf.Name())
					return  CheckFile(zipname+"!"+f.Name(), tf.Name())
				}
				pos += n
			}
		}
		log.Printf("Failed to open %q as zip: %v", zipfile, ziperr)
	}
	defer rdr.Close()
	for _, f := range rdr.File {
		if strings.HasSuffix(f.Name, "jar") || strings.HasSuffix(f.Name, "war") {
			tf, err := ioutil.TempFile(os.TempDir(), "log4j2go-*.jar")
			if err != nil {
				log.Fatal(err)
			}
			defer tf.Close()
			defer os.Remove(tf.Name())
			zrdr, err := f.Open()
			if err != nil {
				log.Fatal("Failed to open embedded zip", err)
			}
			io.Copy(tf, zrdr)

			return CheckFile(zipname+"!"+f.Name, tf.Name())

		}
	}
	if v := JarVersion(rdr.File); v != "" {
		log.Printf("Found %s in %q", v, zipname)
		return v
	}
	return ""
}
