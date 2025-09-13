package main

import (
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"inHack/internal/handler"
	rout "inHack/internal/router"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	logrus.SetFormatter(new(logrus.JSONFormatter))
	if err := initConfig(); err != nil {
		logrus.Fatal("error initializing configs", err)
	}
	if err := godotenv.Load(); err != nil {
		logrus.Fatal("error initializing configs", err)
	}
	//services := service.NewService(repos)
	handlers := handler.NewHandler()
	srv := new(rout.Server)
	go func() {
		if err := srv.Run(viper.GetString("port"), handlers.InitRouter()); err != nil {
			logrus.Fatal("Error in cmd main - ", err)
		}
	}()
	logrus.Println("Todo App upping down")
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit
	logrus.Print("To do App Shutting DOWN")
	if err := srv.Shutdown(); err != nil {
		logrus.Errorf("Error occured on server shutting down: %s", err.Error())
	}

}
func initConfig() error {
	viper.AddConfigPath("configs")
	viper.SetConfigName("config")
	return viper.ReadInConfig()
}
