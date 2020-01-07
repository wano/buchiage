package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/wano/buchiage"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func main() {
	if len(os.Args) < 2 {
		log.Println("must be specify config file.")
		os.Exit(1)
	}

	f, err := os.Open(os.Args[1])
	if err != nil {
		log.Println( err)
		os.Exit(1)
	}

	buf, err := ioutil.ReadAll(f)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	var configList buchiage.ConfigList
	if err := json.Unmarshal(buf, &configList); err != nil {
		log.Println(err)
		os.Exit(1)
	}

	sess, err := session.NewSession()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	s3Client := s3.New(sess)
	runner := buchiage.New(configList.ConfList, s3Client)
	runner.SetFileNameFunc(func(original string, config *buchiage.Config) (string, error) {
		suffix, err := os.Hostname()
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("%s-%s", original, suffix), nil
	})
	done := make(chan os.Signal)
	signal.Notify(done, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := runner.Run(); err != nil {
			log.Println(err)
		}
	}()

	<-done
	if err := runner.Close(); err != nil {
		log.Println(err)
	}

	wg.Wait()
}
