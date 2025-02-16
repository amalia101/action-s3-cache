package main

import (
	"flag"
	"fmt"
	"log"
	"strings"
	"time"
	"os/exec"
)

func main() {
	action := flag.String("action", "", "Action to perform: put, get, delete")
	bucket := flag.String("bucket", "", "S3 bucket name")
	s3Class := flag.String("s3-class", "STANDARD", "S3 storage class")
	key := flag.String("key", "", "Cache key (without .zip)")
	artifacts := flag.String("artifacts", "", "Comma-separated list of artifact paths")

	flag.Parse()

	if *action == "" || *bucket == "" || *key == "" {
		log.Fatal("Missing required arguments: --action, --bucket, and --key are required")
	}

	artifactList := strings.Split(strings.TrimSpace(*artifacts), ",")

	zipKey := fmt.Sprintf("%s.zip", *key)

	switch *action {
	case "put":
		if len(artifactList) == 0 || artifactList[0] == "" {
			log.Fatal("No artifacts provided")
		}
	
		tarCmd := fmt.Sprintf("tar -czvf %s %s", zipKey, strings.Join(artifactList, " "))
		if err := runShellCommand(tarCmd); err != nil {
			log.Fatal("Failed to create tar.gz:", err)
		}
	
		if err := PutObject(zipKey, *bucket, *s3Class); err != nil {
			log.Fatal(err)
		}
	
		today := time.Now().Format("02-01-2006")
		if err := PutTag(zipKey, *bucket, "LastUsedDate", today); err != nil {
			log.Fatal(err)
		}	

	case "get":
		exists, err := ObjectExists(zipKey, *bucket)
		if err != nil {
			log.Fatal(err)
		}
	
		if exists {
			if err := GetObject(zipKey, *bucket); err != nil {
				log.Fatal(err)
			}
	
			extractCmd := fmt.Sprintf("tar -xzvf %s", zipKey)
			if err := runShellCommand(extractCmd); err != nil {
				log.Fatal("Failed to extract tar.gz:", err)
			}
	
			today := time.Now().Format("02-01-2006")
			tag, _ := GetTag(zipKey, *bucket, "LastUsedDate")
	
			if tag != today {
				if err := PutTag(zipKey, *bucket, "LastUsedDate", today); err != nil {
					log.Fatal(err)
				}
			}
		} else {
			log.Printf("No caches found for key: %s", zipKey)
		}	

	case "delete":
		if err := DeleteObject(zipKey, *bucket); err != nil {
			log.Fatal(err)
		}

	default:
		log.Fatalf("Invalid action: %s. Allowed: [put, get, delete]", *action)
	}
}

func runShellCommand(cmdStr string) error {
	cmd := exec.Command("sh", "-c", cmdStr)
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}