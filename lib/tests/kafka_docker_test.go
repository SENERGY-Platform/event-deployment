package tests

import (
	"context"
	"github.com/ory/dockertest"
	"github.com/ory/dockertest/docker"
	"github.com/segmentio/kafka-go"
	"github.com/wvanbergen/kazoo-go"
	"log"
	"net"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
)

func Kafka(pool *dockertest.Pool, ctx context.Context, zookeeperUrl string) (err error) {
	kafkaport, err := getFreePort()
	if err != nil {
		return err
	}
	networks, _ := pool.Client.ListNetworks()
	hostIp := ""
	for _, network := range networks {
		if network.Name == "bridge" {
			hostIp = network.IPAM.Config[0].Gateway
		}
	}
	log.Println("host ip: ", hostIp)
	env := []string{
		"ALLOW_PLAINTEXT_LISTENER=yes",
		"KAFKA_LISTENERS=OUTSIDE://:9092",
		"KAFKA_ADVERTISED_LISTENERS=OUTSIDE://" + hostIp + ":" + strconv.Itoa(kafkaport),
		"KAFKA_LISTENER_SECURITY_PROTOCOL_MAP=OUTSIDE:PLAINTEXT",
		"KAFKA_INTER_BROKER_LISTENER_NAME=OUTSIDE",
		"KAFKA_ZOOKEEPER_CONNECT=" + zookeeperUrl,
	}
	log.Println("start kafka with env ", env)
	container, err := pool.RunWithOptions(&dockertest.RunOptions{Repository: "bitnami/kafka", Tag: "latest", Env: env, PortBindings: map[docker.Port][]docker.PortBinding{
		"9092/tcp": {{HostIP: "", HostPort: strconv.Itoa(kafkaport)}},
	}})
	if err != nil {
		return err
	}
	go func() {
		<-ctx.Done()
		log.Println("DEBUG: remove container " + container.Container.Name)
		container.Close()
	}()
	err = pool.Retry(func() error {
		log.Println("try kafka connection...")
		conn, err := kafka.Dial("tcp", hostIp+":"+strconv.Itoa(kafkaport))
		if err != nil {
			log.Println(err)
			return err
		}
		defer conn.Close()
		return nil
	})
	time.Sleep(5 * time.Second)
	return err
}

func Zookeeper(pool *dockertest.Pool, ctx context.Context) (hostPort string, ipAddress string, err error) {
	zkport, err := getFreePort()
	if err != nil {
		debug.PrintStack()
		return "", "", err
	}
	env := []string{}
	log.Println("start zookeeper on ", zkport)
	container, err := pool.RunWithOptions(&dockertest.RunOptions{Repository: "wurstmeister/zookeeper", Tag: "latest", Env: env, PortBindings: map[docker.Port][]docker.PortBinding{
		"2181/tcp": {{HostIP: "", HostPort: strconv.Itoa(zkport)}},
	}})
	if err != nil {
		debug.PrintStack()
		return "", "", err
	}
	go func() {
		<-ctx.Done()
		log.Println("DEBUG: remove container " + container.Container.Name)
		container.Close()
	}()
	hostPort = strconv.Itoa(zkport)
	err = pool.Retry(func() error {
		log.Println("try zk connection...")
		zookeeper := kazoo.NewConfig()
		zk, chroot := kazoo.ParseConnectionString(container.Container.NetworkSettings.IPAddress)
		zookeeper.Chroot = chroot
		kz, err := kazoo.NewKazoo(zk, zookeeper)
		if err != nil {
			log.Println("kazoo", err)
			return err
		}
		_, err = kz.Brokers()
		if err != nil && strings.TrimSpace(err.Error()) != strings.TrimSpace("zk: node does not exist") {
			log.Println("brokers", err)
			return err
		}
		return nil
	})
	return hostPort, container.Container.NetworkSettings.IPAddress, err
}

func getFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port, nil
}

func Dockerlog(pool *dockertest.Pool, ctx context.Context, repo *dockertest.Resource, name string) {
	out := &LogWriter{logger: log.New(os.Stdout, "["+name+"]", 0)}
	err := pool.Client.Logs(docker.LogsOptions{
		Stdout:       true,
		Stderr:       true,
		Context:      ctx,
		Container:    repo.Container.ID,
		Follow:       true,
		OutputStream: out,
		ErrorStream:  out,
	})
	if err != nil && err != context.Canceled {
		log.Println("DEBUG-ERROR: unable to start docker log", name, err)
	}
}

type LogWriter struct {
	logger *log.Logger
}

func (this *LogWriter) Write(p []byte) (n int, err error) {
	this.logger.Print(string(p))
	return len(p), nil
}
