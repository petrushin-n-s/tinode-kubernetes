package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tinode/chat/pbx"
	"google.golang.org/grpc"
)

type SignupRequest struct {
	UserID   string `json:"userId" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type Client struct {
	pbx.Node_MessageLoopClient
	id int
}

func (c *Client) nextID() string {
	c.id++
	return strconv.Itoa(c.id)
}

func (c *Client) SendHi() {
	hi := &pbx.ClientHi{}
	hi.Id = c.nextID()
	hi.UserAgent = "go_test/1.0"
	hi.Ver = "0.22.11"
	hi.Lang = "EN"

	clientMessage := &pbx.ClientMsg{Message: &pbx.ClientMsg_Hi{hi}}
	err := c.Send(clientMessage)

	if err != nil {
		log.Fatal("error sending message ", err)
	}

	serverMsg, err := c.Recv()
	if err != nil {
		log.Fatal(err)
	}
	log.Println(serverMsg)
}

func (c *Client) SendAcc(query SignupRequest) (string, error) {
	var cred []*pbx.ClientCred
	cred = append(cred, &pbx.ClientCred{Method: "email", Value: query.Email})

	secret := fmt.Sprintf("%s:%s", query.Username, query.Password)
	public := fmt.Sprintf("{\"fn\":\"%s\"}", query.UserID)
	accMsg := &pbx.ClientMsg{Message: &pbx.ClientMsg_Acc{
		Acc: &pbx.ClientAcc{
			Id:     c.nextID(),
			UserId: "new", Scheme: "basic",
			Secret: []byte(secret),
			Login:  true,
			Desc:   &pbx.SetDesc{Public: []byte(public)},
			Cred:   cred},
	}}

	if err := c.Send(accMsg); err != nil {
		return "", fmt.Errorf("error sending message %w", err)
	}

	serverMsg, err := c.Recv()
	if err != nil {
		log.Fatal(err)
	}
	ctrl := serverMsg.GetCtrl()
	if ctrl.Code != 200 {
		return "", fmt.Errorf("error while creating account %v", ctrl)
	}

	return string(ctrl.Params["user"]), nil

}

func NewClient(conn *grpc.ClientConn) Client {
	c := pbx.NewNodeClient(conn)
	loop, err := c.MessageLoop(context.Background())

	if err != nil {
		log.Fatal("Error calling", err)
	}

	return Client{
		Node_MessageLoopClient: loop,
	}
}

const HOST = "tinode-service:16060"

func main() {
	conn, err := grpc.Dial(HOST, grpc.WithInsecure())
	if err != nil {
		log.Fatal("Error dialing", err)
	}

	client := NewClient(conn)
	r := gin.Default()

	r.POST("/signup", func(c *gin.Context) {
		var request SignupRequest

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		fmt.Printf("Received signup request: %+v\n", request)

		client.SendHi()
		userID, err := client.SendAcc(request)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "User signed up successfully!", "userID": userID,
		})
	})

	_ = r.Run(":8080")
}
