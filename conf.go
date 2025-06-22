package main

import (
	"encoding/json"
	"os"
)

type Conf struct {
	Ip         string        `json:"ip"`
	Port       int           `json:"port"`
	Peers      []interface{} `json:"peers"`
	Dir        string        `json:"dir"`
	Passwd     string        `json:"passwd"`
	MaxDB      int           `json:"max_db"`
	ShardCount int           `json:"shard_count"`
	AOFFile    string        `json:"aof_file"`
	AOFFsync   string        `json:"aof_fsync"`
}

func GetConf() *Conf {
	bs, err := os.ReadFile(ConfPath)
	HandleErr(err)
	conf := &Conf{}
	err = json.Unmarshal(bs, conf)
	HandleErr(err)
	return conf
}
