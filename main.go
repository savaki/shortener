package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/hashicorp/golang-lru"
	"gopkg.in/urfave/cli.v1"
)

var opts struct {
	Port     int
	Bucket   string
	Codebase string
}

func main() {
	app := cli.NewApp()
	app.Flags = []cli.Flag{
		cli.IntFlag{
			Name:        "port",
			Usage:       "port to listen to; defaults to 3000",
			Value:       3000,
			EnvVar:      "PORT",
			Destination: &opts.Port,
		},
		cli.StringFlag{
			Name:        "bucket",
			Usage:       "s3 bucket to save shortened urls to",
			Value:       "lovingly-shortened",
			EnvVar:      "S3_BUCKET",
			Destination: &opts.Bucket,
		},
		cli.StringFlag{
			Name:        "codebase",
			Usage:       "cloudfront codebase",
			Value:       "http://d1gssjv4jvarn7.cloudfront.net",
			EnvVar:      "CODEBASE",
			Destination: &opts.Codebase,
		},
	}
	app.Action = listenAndServe
	app.Run(os.Args)
}

type server struct {
	cache    *lru.Cache
	s3       *s3.S3
	bucket   string
	codebase string
}

func (s *server) decode(c *gin.Context) {
	if c.Request.Method != http.MethodGet {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	// look inside lru for recently used key
	key := c.Request.URL.Path
	if v, ok := s.cache.Get(key); ok {
		c.Redirect(http.StatusTemporaryRedirect, v.(string))
		return
	}
	fmt.Println("cache miss,", key)

	// fetch content from cloudfront
	//
	req, err := http.NewRequest(http.MethodGet, s.codebase+key, nil)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	req = req.WithContext(c.Request.Context())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err)
		statusCode := http.StatusInternalServerError
		if resp != nil {
			statusCode = resp.StatusCode
		}
		c.AbortWithError(statusCode, err)
		return
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// ignore file not found
	//
	location := string(data)
	if strings.HasPrefix(location, `<?xml`) {
		c.String(http.StatusNotFound, "url not found")
		return
	}

	// update cache
	//
	fmt.Printf("adding to cache '%v' => '%v'\n", key, location)
	s.cache.ContainsOrAdd(key, location)

	// and redirect
	//
	c.Redirect(http.StatusTemporaryRedirect, location)
}

func (s *server) register(c *gin.Context) {
	key := c.PostForm("key")
	url := c.PostForm("url")

	_, err := s.s3.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Body:   strings.NewReader(url),
	})
	if err != nil {
		fmt.Println(err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	s.cache.Add(key, url)
	c.String(http.StatusOK, "ok")
}

func listenAndServe(_ *cli.Context) error {
	region := os.Getenv("AWS_DEFAULT_REGION")
	s, err := session.NewSession(&aws.Config{Region: aws.String(region)})
	if err != nil {
		log.Fatalln(err)
	}

	cache, err := lru.New(1024)
	if err != nil {
		log.Fatalln(err)
	}

	server := &server{
		cache:    cache,
		s3:       s3.New(s),
		bucket:   opts.Bucket,
		codebase: opts.Codebase,
	}

	router := gin.New()
	router.POST("/register", server.register)
	router.NoRoute(server.decode)
	http.ListenAndServe(fmt.Sprintf(":%d", opts.Port), router)

	return nil
}
