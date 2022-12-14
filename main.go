package main

import (
	"context"	
	"fmt"	
	"net/http"
	"os"
	"os/signal"
	"sync"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	"encoding/json"
	"io/ioutil"	
	// "time"
)

var Donewait sync.WaitGroup

type User struct {
	Id string `json:"id"`
	Email string `json:"email"`
	Valid bool `json:"valid"`
}

type JobData struct {
	Count              	int
	ID                 	int		
	Subject							string
	Contents						string
	JobType						 	int	
}

var JobQueue chan JobData
var count int

func ReadConfig(filepath string) {
	viper.SetConfigFile(filepath)
	viper.ReadInConfig()
}

type SendEmailData struct {	
	Email             string    `form:"email" json:"email" xml:"email"  binding:"required"` 
	Subject           string 		`form:"subject" json:"subject" xml:"subject"  binding:"required"` 
	Contents          string    `form:"contents" json:"contents" xml:"contents"  binding:"required"`	
}

func SendEmail(c *gin.Context) {

	var jsonData SendEmailData
	if err := c.ShouldBindJSON(&jsonData); err != nil {
		fmt.Println(jsonData)
		c.JSON(http.StatusBadRequest, gin.H{"ct": 1, "message": err.Error()})
		return
	}

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		fmt.Println(err)
	}
	client_ses := ses.NewFromConfig(cfg)

	input := &ses.SendEmailInput{
		Destination: &types.Destination{
				CcAddresses: []string{
				},
				ToAddresses: []string{
					jsonData.Email,
				},
		},
		Message: &types.Message{
				Body: &types.Body{
						Html: &types.Content{
								Charset: aws.String("UTF-8"),
								Data:    aws.String(jsonData.Contents),
						},						
				},
				Subject: &types.Content{
						Charset: aws.String("UTF-8"),
						Data:    aws.String(jsonData.Subject),
				},
		},
		Source: aws.String(viper.GetString(("from-address"))),				
	}

	output, err := client_ses.SendEmail(context.TODO(), input)
	if err != nil {
		fmt.Println(err)		
		c.JSON(http.StatusBadRequest, gin.H{"ct": 1, "message": err.Error()})
		return
	}
	fmt.Println(output)
	c.JSON(http.StatusOK, gin.H{"ct": 0})
}

type SendAllData struct {		
	Subject           string 		`form:"subject" json:"subject" xml:"subject"  binding:"required"` 
	Contents          string    `form:"contents" json:"contents" xml:"contents"  binding:"required"`	
}

func SendAll(c *gin.Context) {
	var jsonData SendAllData
	if err := c.ShouldBindJSON(&jsonData); err != nil {
		fmt.Println(jsonData)
		c.JSON(http.StatusBadRequest, gin.H{"ct": 1, "message": err.Error()})
		return
	}

	count++
	var job JobData
	job.Count = count
	job.Subject = jsonData.Subject
	job.Contents = jsonData.Contents
	job.JobType = 0

	JobQueue <- job

	c.JSON(http.StatusOK, gin.H{"ct": 0})
}

func SendAllCertification(c *gin.Context) {
	
	count++
	var job JobData
	job.Count = count	
	job.JobType = 1
	JobQueue <- job
	c.JSON(http.StatusOK, gin.H{"ct": 0})
}

func SendTest(c *gin.Context) {

}

func (job *JobData) Process() error {

	if job.JobType == 0 {
		fmt.Println("jobtype0")
		url := fmt.Sprintf("%v/user/email", viper.GetString("aif-api-url"))
		// url := fmt.Sprintf("%v/user/aif-email", viper.GetString("aif-api-url"))
		resp, err := http.Get(url)
		if err != nil {			
			return err
		}
		defer resp.Body.Close()
		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {			
			return err
		}
		userList := make([]User,0)
		if err := json.Unmarshal(respBody, &userList); err != nil {
			fmt.Println(err)
			return err
		}	

		cfg, err := config.LoadDefaultConfig(context.TODO())
		if err != nil {
			fmt.Println(err)
		}
		client_ses := ses.NewFromConfig(cfg)

		for _,item := range userList {
			input := &ses.SendEmailInput{
				Destination: &types.Destination{
						CcAddresses: []string{
						},
						ToAddresses: []string{
							item.Email,
						},
				},
				Message: &types.Message{
						Body: &types.Body{
								Html: &types.Content{
										Charset: aws.String("UTF-8"),
										Data:    aws.String(job.Contents),
								},						
						},
						Subject: &types.Content{
								Charset: aws.String("UTF-8"),
								Data:    aws.String(job.Subject),
						},
				},
				Source: aws.String(viper.GetString(("from-address"))),				
			}
		
			_, err := client_ses.SendEmail(context.TODO(), input)
			if err != nil {
				fmt.Println(err.Error())									
			}
			//fmt.Println(output)
			//time.Sleep(time.Millisecond * 50)
		}

	} else if job.JobType == 1 {
		url := fmt.Sprintf("%v/user/email", viper.GetString("aif-api-url"))
		//url := fmt.Sprintf("%v/user/aif-email", viper.GetString("aif-api-url"))
		resp, err := http.Get(url)
		if err != nil {			
			return err
		}
		defer resp.Body.Close()
		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {			
			return err
		}
		userList := make([]User,0)
		if err := json.Unmarshal(respBody, &userList); err != nil {
			fmt.Println(err)
			return err
		}	

		cfg, err := config.LoadDefaultConfig(context.TODO())
		if err != nil {
			fmt.Println(err)
		}
		
		client_ses := ses.NewFromConfig(cfg)
		subject := "[?????????????????????] ????????? ?????? ??????"
		

		for i ,item := range userList {
			link := viper.GetString("aif-api-url") + "/finish-confirmation?email=" + item.Email + "&id=" + item.Id
			contents := `<h4>???????????????.</h4><h4>????????????????????? ???????????? ?????????????????? ???????????? ??????????????????.</h4>
			<h4>?????? '?????? ??????' ????????? ???????????? ????????? ????????? ?????????.</h4><a style="font-size:14px;" href="` + link + `">????????????</a>	<h4>???????????????.</h4><h4>--------</h4><h4>??? ????????? ?????????????????????.</h4>`
			input := &ses.SendEmailInput{
				Destination: &types.Destination{
						CcAddresses: []string{
						},
						ToAddresses: []string{
							item.Email,
						},
				},
				Message: &types.Message{
						Body: &types.Body{
								Html: &types.Content{
										Charset: aws.String("UTF-8"),
										Data:    aws.String(contents),
								},						
						},
						Subject: &types.Content{
								Charset: aws.String("UTF-8"),
								Data:    aws.String(subject),
						},
				},
				Source: aws.String(viper.GetString(("from-address"))),				
			}
		
			_, err := client_ses.SendEmail(context.TODO(), input)
			if err != nil {
				fmt.Println(err.Error())									
				fmt.Println(item.Email)
			}
			fmt.Println(i)
			//fmt.Println(output)
			//time.Sleep(time.Millisecond * 50)
		}
	}

	return nil
}

func DoWork(id int, jobs <-chan JobData, done *sync.WaitGroup, quit <-chan bool) {
	fmt.Println("dowork")
	done.Add(1)
	for {
		select {
		case job := <-jobs:
			job.ID = id
			if err := job.Process(); err != nil {
				fmt.Println(err)
				
				// if job.JobType == 0 {				
				// 	go UpdateSubmission(job.SubmissionIDPublic,job.SubmissionIDTotal,-1, -1,  err.Error())
				// } else if job.JobType == 1 {
				// 	go UpdateDemoApiError(job.ModelServiceID, err.Error())
				// }
				
			}
		case <-quit:
			fmt.Printf("dowork terminated : %d \n", id)
			done.Done()
			return
		}
	}
}

func main() {
	ReadConfig("./config.json")
	count = 0
	JobQueue = make(chan JobData, 100)
	Donewait := sync.WaitGroup{}
	
	chans := make([]chan bool, viper.GetInt("max-worker"))
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, os.Kill)

	fmt.Println(viper.GetInt("max-worker"))

	for w := 0; w <= viper.GetInt("max-worker")-1; w++ {
		chans[w] = make(chan bool)
		go DoWork(w, JobQueue, &Donewait, chans[w])
	}

	router := gin.Default()
	router.Use(cors.Default())
	router.POST("/send-email", SendEmail)
	router.POST("/send-all", SendAll)
	router.POST("/send-all-cert", SendAllCertification)

	srv := &http.Server{
		Addr:    ":9876",
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("http error %v \n", err)
			signalChan <- os.Interrupt
		}
	}()
	fmt.Println("running...")
	<-signalChan
	
	fmt.Println("terminating...")
	
	for w := 0; w <= viper.GetInt("max-worker")-1; w++ { // ?????????, done-terminate
		chans[w] <- true
	}
	fmt.Println("waiting...")
	Donewait.Wait()
	fmt.Println("done all...")

}


