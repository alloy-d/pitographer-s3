package main

import (
	"encoding/json"
	"flag"
	"log"
	"os/exec"
	"sort"
	"time"

	"launchpad.net/goamz/aws"
	"launchpad.net/goamz/s3"
)

var bucket *s3.Bucket

var bucketName = flag.String("bucket", "bitjester.co", "the name of the bucket to upload into")
var prefix = flag.String("prefix", "diacam/", "a string with which to prefix names of uploads")
var updateManifest = flag.Bool("update-manifest", false, "add uploaded file as an entry in ${prefix}MANIFEST.json")

func init() {
	auth, err := aws.EnvAuth()
	if err != nil {
		log.Panicln("couldn't get AWS auth credentials: " + err.Error())
	}

	region := aws.USEast

	conn := s3.New(auth, region)
	bucket = conn.Bucket(*bucketName)
}

func main() {
	flag.Parse()
	
	filename := time.Now().Format(*prefix + "2006-01-02T15:04:05-0700.jpg")
	capture := exec.Command("raspistill", "-o", "-")

	image, err := capture.Output()
	if err != nil {
		log.Panicln("couldn't capture an image: " + err.Error())
	}

	err = upload(filename, image)
	if err != nil {
		log.Panicln("couldn't upload image to S3: " + err.Error())
	}

	log.Printf("uploaded %s\n", filename)

	if *updateManifest {
		writeManifest()
	}
}

func upload(filename string, contents []byte) (err error) {
	err = bucket.Put(filename, contents, "image/jpeg", s3.PublicRead)
	return
}

func writeManifest() (err error) {
	resp, err := bucket.List(*prefix + "201", "", *prefix + "201", 1000)
	if err != nil {
		return
	}

	keys := make([]string, 0)
	for _, key := range resp.Contents {
		keys = append(keys, key.Key)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(keys)))

	log.Println(keys)

	json, err := json.Marshal(keys)
	if err != nil {
		return
	}

	bucket.Put(*prefix + "MANIFEST.json", json, "application/json", s3.PublicRead)

	return
}
