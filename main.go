package main

import (
	"context"
	"log"
	"strings"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"

	proxymanService "github.com/v2fly/v2ray-core/v4/app/proxyman/command"
	statsService "github.com/v2fly/v2ray-core/v4/app/stats/command"
	"github.com/v2fly/v2ray-core/v4/common/protocol"
	"github.com/v2fly/v2ray-core/v4/common/serial"
	"github.com/v2fly/v2ray-core/v4/proxy/vmess"
)

func addUser(email string, uuid string) error {
	conn, openErr := grpc.Dial("127.0.0.1:10085", grpc.WithInsecure())
	if openErr != nil {
		return openErr
	}
	client := proxymanService.NewHandlerServiceClient(conn)
	_, err := client.AlterInbound(context.Background(), &proxymanService.AlterInboundRequest{
		Tag: "proxy",
		Operation: serial.ToTypedMessage(&proxymanService.AddUserOperation{
			User: &protocol.User{
				Level: 0,
				Email: email,
				Account: serial.ToTypedMessage(&vmess.Account{
					Id:               uuid,
					AlterId:          0,
					SecuritySettings: &protocol.SecurityConfig{Type: protocol.SecurityType_AUTO},
				}),
			},
		}),
	})
	if closeErr := conn.Close(); closeErr != nil {
		log.Println(closeErr)
	}
	if err != nil {
		return err
	}
	return nil
}
func removeUser(email string) error {
	conn, openErr := grpc.Dial("127.0.0.1:10085", grpc.WithInsecure())
	if openErr != nil {
		return openErr
	}
	client := proxymanService.NewHandlerServiceClient(conn)
	_, err := client.AlterInbound(context.Background(), &proxymanService.AlterInboundRequest{
		Tag: "proxy",
		Operation: serial.ToTypedMessage(&proxymanService.RemoveUserOperation{
			Email: email,
		}),
	})
	if closeErr := conn.Close(); closeErr != nil {
		log.Println(closeErr)
	}
	if err != nil {
		return err
	}
	return nil
}
func queryUserTraffic(email string) ([]*statsService.Stat, error) {
	conn, openErr := grpc.Dial("127.0.0.1:10085", grpc.WithInsecure())
	if openErr != nil {
		return nil, openErr
	}
	client := statsService.NewStatsServiceClient(conn)
	resp, err := client.QueryStats(context.Background(), &statsService.QueryStatsRequest{
		Pattern: "user>>>" + email,
		Reset_:  false,
	})
	if closeErr := conn.Close(); closeErr != nil {
		log.Println(closeErr)
	}
	if err != nil {
		return nil, err
	}
	return resp.GetStat(), nil
}
func queryTraffic() ([]*statsService.Stat, error) {
	conn, openErr := grpc.Dial("127.0.0.1:10085", grpc.WithInsecure())
	if openErr != nil {
		return nil, openErr
	}
	client := statsService.NewStatsServiceClient(conn)
	resp, err := client.QueryStats(context.Background(), &statsService.QueryStatsRequest{
		Pattern: "user>>>",
		Reset_:  false,
	})
	if closeErr := conn.Close(); closeErr != nil {
		log.Println(closeErr)
	}
	if err != nil {
		return nil, err
	}
	return resp.GetStat(), nil
}
func resetUserTraffic(email string) ([]*statsService.Stat, error) {
	conn, openErr := grpc.Dial("127.0.0.1:10085", grpc.WithInsecure())
	if openErr != nil {
		return nil, openErr
	}
	client := statsService.NewStatsServiceClient(conn)
	resp, err := client.QueryStats(context.Background(), &statsService.QueryStatsRequest{
		Pattern: "user>>>" + email,
		Reset_:  true,
	})
	if closeErr := conn.Close(); closeErr != nil {
		log.Println(closeErr)
	}
	if err != nil {
		return nil, err
	}
	return resp.GetStat(), nil
}
func resetTraffic() ([]*statsService.Stat, error) {
	conn, openErr := grpc.Dial("127.0.0.1:10085", grpc.WithInsecure())
	if openErr != nil {
		return nil, openErr
	}
	client := statsService.NewStatsServiceClient(conn)
	resp, err := client.QueryStats(context.Background(), &statsService.QueryStatsRequest{
		Pattern: "user>>>",
		Reset_:  true,
	})
	if closeErr := conn.Close(); closeErr != nil {
		log.Println(closeErr)
	}
	if err != nil {
		return nil, err
	}
	return resp.GetStat(), nil
}
func main() {
	router := gin.Default()
	router.GET("/add-user", func(c *gin.Context) {
		query := c.Request.URL.Query()
		email := query.Get("email")
		uuid := query.Get("uuid")
		if len(email) == 0 || len(uuid) != 36 {
			c.Status(400)
			return
		}
		err := addUser(email, uuid)
		if err != nil {
			log.Println(err)
			if strings.Contains(err.Error(), "User "+email+" already exists.") {
				c.Status(409)
				return
			}
			c.Status(500)
			return
		}
		c.Status(200)
	})
	router.GET("/remove-user", func(c *gin.Context) {
		query := c.Request.URL.Query()
		email := query.Get("email")
		if len(email) == 0 {
			c.Status(400)
			return
		}
		err := removeUser(email)
		if err != nil {
			log.Println(err)
			if strings.Contains(err.Error(), "User "+email+" not found.") {
				c.Status(404)
				return
			}
			c.Status(500)
			return
		}
		c.Status(200)
	})
	router.GET("/query-user-traffic", func(c *gin.Context) {
		query := c.Request.URL.Query()
		email := query.Get("email")
		if len(email) == 0 {
			c.Status(400)
			return
		}
		stats, err := queryUserTraffic(email)
		if err != nil {
			log.Println(err)
			c.Status(500)
			return
		}
		c.JSON(200, stats)
	})
	router.GET("/query-traffic", func(c *gin.Context) {
		stats, err := queryTraffic()
		if err != nil {
			log.Println(err)
			c.Status(500)
			return
		}
		c.JSON(200, stats)
	})
	router.GET("/reset-user-traffic", func(c *gin.Context) {
		query := c.Request.URL.Query()
		email := query.Get("email")
		if len(email) == 0 {
			c.Status(400)
			return
		}
		stats, err := resetUserTraffic(email)
		if err != nil {
			log.Println(err)
			c.Status(500)
			return
		}
		c.JSON(200, stats)
	})
	router.GET("/reset-traffic", func(c *gin.Context) {
		stats, err := resetTraffic()
		if err != nil {
			log.Println(err)
			c.Status(500)
			return
		}
		c.JSON(200, stats)
	})
	router.Run("127.0.0.1:10087")
}
