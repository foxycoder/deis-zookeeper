package tests

import (
  "fmt"
  "testing"
  "time"

  "github.com/deis/deis/tests/dockercli"
  "github.com/deis/deis/tests/etcdutils"
  "github.com/deis/deis/tests/utils"
)

func TestZookeeper(t *testing.T) {
  var err error
  setkeys := []string{
    "/deis/zookeeper/host",
    "/deis/zookeeper/port",
    "/deis/solrcloud/host",
    "/deis/solrcloud/port"
  }
  setdir := []string{
    "/deis/zookeeper",
    "/deis/solrcloud"
  }
  tag, etcdPort := utils.BuildTag(), utils.RandomPort()
  etcdName := "deis-etcd-" + tag
  cli, stdout, stdoutPipe := dockercli.NewClient()
  dockercli.RunTestEtcd(t, etcdName, etcdPort)
  defer cli.CmdRm("-f", etcdName)
  handler := etcdutils.InitEtcd(setdir, setkeys, etcdPort)
  etcdutils.PublishEtcd(t, handler)
  host, port := utils.HostAddress(), utils.RandomPort()
  fmt.Printf("--- Run deis/zookeeper:%s at %s:%s\n", tag, host, port)
  name := "deis-zookeeper-" + tag
  go func() {
    _ = cli.CmdRm("-f", name)
    err = dockercli.RunContainer(cli,
      "--name", name,
      "--rm",
      "-p", port+":80",
      "-p", utils.RandomPort()+":2222",
      "-e", "EXTERNAL_PORT="+port,
      "-e", "HOST="+host,
      "-e", "ETCD_PORT="+etcdPort,
      "deis/zookeeper:"+tag)
  }()
  dockercli.PrintToStdout(t, stdout, stdoutPipe, "deis-zookeeper running")
  if err != nil {
    t.Fatal(err)
  }
  // FIXME: Zookeeper needs a couple seconds to wake up here
  // FIXME: Wait until etcd keys are published
  time.Sleep(5000 * time.Millisecond)
  dockercli.DeisServiceTest(t, name, port, "tcp")
  zookeeperKeyPrefix := "/deis/zookeeper/" + host
  etcdutils.VerifyEtcdValue(t, zookeeperKeyPrefix+"/host", host, etcdPort)
  etcdutils.VerifyEtcdValue(t, zookeeperKeyPrefix+"/port", port, etcdPort)
  _ = cli.CmdRm("-f", name)
}
